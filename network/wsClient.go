package network

import (
	"bytes"
	"context"
	"fmt"
	"github.com/constellation39/majsoul/message"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	msgTypeNotify   uint8 = 1 // 通知
	msgTypeRequest  uint8 = 2 // 请求
	msgTypeResponse uint8 = 3 // 回复
)

type reply struct {
	out   proto.Message
	wait  chan struct{}
	index uint8
}

type WsClient struct {
	conn               *websocket.Conn
	ConnAddress        string
	DialOptions        websocket.DialOptions
	messageIndex       uint32
	requestResponseMap sync.Map // map[uint8]*reply
	notify             chan *message.Wrapper
	reconnectHandler   func()
}

func NewWsClient(connAddress string, dialOptions websocket.DialOptions, reconnectHandler func()) *WsClient {
	return &WsClient{
		conn:               nil,
		ConnAddress:        connAddress,
		DialOptions:        dialOptions,
		messageIndex:       0,
		requestResponseMap: sync.Map{},
		notify:             make(chan *message.Wrapper, 64),
		reconnectHandler:   reconnectHandler,
	}
}

func (client *WsClient) Connect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	conn, _, err := websocket.Dial(ctx, client.ConnAddress, &client.DialOptions)
	if err != nil {
		return fmt.Errorf("majsoul ws failed to dial error %v", err)
	}
	conn.SetReadLimit(1048576)
	client.conn = conn

	go client.readLoop()
	return nil
}

func (client *WsClient) Receive() <-chan *message.Wrapper {
	return client.notify
}

func (client *WsClient) Close() error {
	if client.conn != nil {
		if err := client.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			return fmt.Errorf("websocket conn close error: %w", err)
		}
	}
	client.conn = nil
	client.messageIndex = 0
	client.requestResponseMap = sync.Map{}
	return nil
}

func (client *WsClient) readLoop() {
	for {
		var msgType websocket.MessageType
		var payload []byte
		var err error
		{
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, time.Minute)
			msgType, payload, err = client.conn.Read(ctx)
			cancel()
		}
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				break
			}
			for {
				ctx := context.Background()
				ctx, cancel := context.WithTimeout(ctx, time.Second*5)
				err = client.Connect(ctx)
				cancel()
				if err == nil {
					if client.reconnectHandler != nil {
						client.reconnectHandler()
					}
					break
				}
				time.Sleep(time.Second * 5)
			}
			continue
		}
		if msgType != websocket.MessageBinary {
			panic(fmt.Sprintf("read message type != %d, type == %d", websocket.MessageBinary, msgType))
		}
		if len(payload) == 0 {
			panic(fmt.Sprintf("read message failed payload is nil"))
		}
		switch payload[0] {
		case msgTypeNotify:
			client.handleNotify(payload)
		case msgTypeResponse:
			client.handleResponse(payload)
		default:
			panic(fmt.Sprintf("read message match unknown message (%d)", payload[0]))
		}
	}
}

func (client *WsClient) handleNotify(msg []byte) {
	wrapper := new(message.Wrapper)

	err := proto.Unmarshal(msg[1:], wrapper)
	if err != nil {
		panic(fmt.Sprintf("notify messages unmarshal error %v", err))
	}

	select {
	case client.notify <- wrapper:
	default:
		panic(" notify channel is full")
	}
}

func (client *WsClient) handleResponse(msg []byte) {
	index := (msg[2] << 7) + msg[1]

	response, ok := client.requestResponseMap.Load(index)
	if !ok {
		return
	}

	r, ok := response.(*reply)
	if !ok {
		panic(fmt.Sprintf("response type (type = %+v) not proto.Message", r))
	}

	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		panic(fmt.Sprintf("response message unmarshal failed error %v", err))
	}

	err = proto.Unmarshal(wrapper.Data, r.out)
	if err != nil {
		panic(fmt.Sprintf("response message unmarshal failed error %v", err))
	}

	close(r.wait)
}

func (client *WsClient) Invoke(ctx context.Context, method string, in interface{}, out interface{}, _ ...grpc.CallOption) error {
	tokens := strings.Split(method, "/")
	api := strings.Join(tokens, ".")

	r, err := client.sendMsg(ctx, api, in.(proto.Message))
	if err != nil {
		return err
	}
	r.out = out.(proto.Message)

	return client.recvMsg(ctx, r)
}

func (client *WsClient) sendMsg(ctx context.Context, api string, in proto.Message) (_ *reply, err error) {
	var body []byte

	body, err = proto.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("majsoul ws failed to marshal message: %v, error: %w", in, err)
	}

	wrapper := &message.Wrapper{
		Name: api,
		Data: body,
	}

	body, err = proto.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("majsoul ws failed to marshal wrapper message: %v, error: %w", wrapper, err)
	}

	index := atomic.LoadUint32(&client.messageIndex)
	if index == 255 {
		index = 0
		atomic.StoreUint32(&client.messageIndex, index)
	} else {
		atomic.AddUint32(&client.messageIndex, 1)
	}

	indexUint8 := uint8(index)

	buff := new(bytes.Buffer)
	buff.WriteByte(msgTypeRequest)
	buff.WriteByte(indexUint8 - (indexUint8 >> 7 << 7))
	buff.WriteByte(indexUint8 >> 7)
	buff.Write(body)

	err = client.conn.Write(ctx, websocket.MessageBinary, buff.Bytes())

	if err != nil {
		return
	}

	r := &reply{
		out:   nil,
		wait:  make(chan struct{}),
		index: indexUint8,
	}

	if _, ok := client.requestResponseMap.LoadOrStore(r.index, r); ok {
		return nil, fmt.Errorf("majsoul ws request index %d already exists", r.index)
	}

	return r, nil
}

func (client *WsClient) recvMsg(ctx context.Context, reply *reply) error {
	defer client.requestResponseMap.Delete(reply.index)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-reply.wait:
	}
	return nil
}

func (client *WsClient) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("implement me")
}
