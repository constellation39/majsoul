package network

import (
	"bytes"
	"context"
	"fmt"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	msgTypeNotify   uint8 = 1
	msgTypeRequest  uint8 = 2
	msgTypeResponse uint8 = 3
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
	ReconnectHandler   func()
}

// NewWsClient creates a new WebSocket client with the specified connection address and dial options.
func NewWsClient(connAddress string, dialOptions websocket.DialOptions) *WsClient {
	return &WsClient{
		conn:               nil,
		ConnAddress:        connAddress,
		DialOptions:        dialOptions,
		messageIndex:       0,
		requestResponseMap: sync.Map{},
		notify:             make(chan *message.Wrapper, 64),
		ReconnectHandler:   nil,
	}
}

// Connect establishes a connection to the WebSocket server. It returns an error if the connection cannot be established.
func (client *WsClient) Connect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	conn, _, err := websocket.Dial(ctx, client.ConnAddress, &client.DialOptions)
	if err != nil {
		return fmt.Errorf("majsoul ws failed to dial, error: %v", err)
	}
	conn.SetReadLimit(1048576)
	client.conn = conn

	go client.readLoop()
	return nil
}

// Receive returns a channel that can be used to receive messages from the WebSocket server.
func (client *WsClient) Receive() <-chan *message.Wrapper {
	return client.notify
}

// Close closes the WebSocket connection. It returns an error if the connection cannot be closed properly.
func (client *WsClient) Close() error {
	if client.conn != nil {
		if err := client.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			return fmt.Errorf("error while closing websocket connection: %v", err)
		}
	}
	client.conn = nil
	client.messageIndex = 0
	client.requestResponseMap = sync.Map{}
	return nil
}

// readLoop continually reads messages from the WebSocket server and handles them according to their type.
func (client *WsClient) readLoop() {
	for {
		var msgType websocket.MessageType
		var payload []byte
		var err error
		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			msgType, payload, err = client.conn.Read(ctx)
			cancel()
		}
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				break
			}
			for {
				{
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					err = client.Connect(ctx)
					cancel()
				}
				if err == nil {
					if client.ReconnectHandler != nil {
						client.ReconnectHandler()
					}
					break
				}
				time.Sleep(time.Second * 5)
			}
			return
		}
		if msgType != websocket.MessageBinary {
			logger.Error("expected message type is not Binary", zap.Int("type", int(msgType)))
			continue
		}
		if len(payload) == 0 {
			logger.Error("read message failed, payload is nil")
			continue
		}
		switch payload[0] {
		case msgTypeNotify:
			client.handleNotify(payload)
		case msgTypeResponse:
			client.handleResponse(payload)
		default:
			logger.Error("read message matched unknown message type", zap.Int("type", int(payload[0])))
		}
	}
}

// handleNotify handles notify messages received from the WebSocket server.
func (client *WsClient) handleNotify(msg []byte) {
	wrapper := new(message.Wrapper)

	err := proto.Unmarshal(msg[1:], wrapper)
	if err != nil {
		logger.Error("error while unmarshalling notify messages", zap.String("msg", string(msg[1:])), zap.Error(err))
		return
	}

	select {
	case client.notify <- wrapper:
	default:
		logger.Panic("notify channel is full, unable to push new message")
	}
}

// handleResponse handles response messages received from the WebSocket server.
func (client *WsClient) handleResponse(msg []byte) {
	index := (msg[2] << 7) + msg[1]

	response, ok := client.requestResponseMap.Load(index)
	if !ok {
		return
	}

	r, ok := response.(*reply)
	if !ok {
		logger.Error("response type is not proto.Message", zap.Reflect("response", response))
	}

	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		logger.Error("error while unmarshal response message", zap.String("msg", string(msg[3:])))

	}

	err = proto.Unmarshal(wrapper.Data, r.out)
	if err != nil {
		logger.Error("error while unmarshal wrapper data", zap.String("data", string(wrapper.Data)))
	}

	close(r.wait)
}

// Invoke sends a request to the WebSocket server and waits for the response.
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

// sendMsg sends a message to the WebSocket server. It returns an error if the message cannot be sent.
func (client *WsClient) sendMsg(ctx context.Context, api string, in proto.Message) (_ *reply, err error) {
	var body []byte

	body, err = proto.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ws message: %v, error: %w", in, err)
	}

	wrapper := &message.Wrapper{
		Name: api,
		Data: body,
	}

	body, err = proto.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ws wrapper message: %v, error: %w", wrapper, err)
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
		return nil, fmt.Errorf("ws request with index %d already exists", r.index)
	}

	return r, nil
}

// recvMsg waits for a response message from the WebSocket server. It returns an error if the response is not received within the context's deadline.
func (client *WsClient) recvMsg(ctx context.Context, reply *reply) error {
	defer client.requestResponseMap.Delete(reply.index)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-reply.wait:
	}
	return nil
}

// NewStream is not implemented in this client.
func (client *WsClient) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("method not implemented")
}
