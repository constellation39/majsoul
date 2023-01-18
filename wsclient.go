package majsoul

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/constellation39/majsoul/logger"
	"go.uber.org/zap"
	"net/http"
	"nhooyr.io/websocket"
	"strings"
	"sync"
	"time"

	"github.com/constellation39/majsoul/message"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

var (
	ErrNotConnected = errors.New("websocket not connected")
)

type wsConfig struct {
	connAddr          string
	proxyAddr         string
	HTTPHeader        http.Header
	Reconnect         bool
	ReconnectInterval time.Duration
	ReconnectNumber   int
}

type wsClient struct {
	*wsConfig
	curReconnectNumber int

	conn *websocket.Conn
	open bool

	msgIndex uint8
	replyMap sync.Map // map[uint8]*Reply
	notify   chan proto.Message

	mu           sync.Mutex
	HandleReConn func()
}

type Reply struct {
	out      proto.Message
	wait     chan struct{}
	msgIndex uint8
	timeOut  bool
}

func newWsClient(config *wsConfig) *wsClient {
	return &wsClient{
		wsConfig: config,
		conn:     nil,
		open:     false,
		msgIndex: 0,
		replyMap: sync.Map{},
		notify:   make(chan proto.Message, 32),
		mu:       sync.Mutex{},
	}
}

func (client *wsClient) Close() {
	client.mu.Lock()
	if client.conn != nil {
		err := client.conn.Close(websocket.StatusNormalClosure, "")
		if err != nil {
			logger.Error("wsClient.conn.Close()", zap.Error(err))
		}
	}
	client.open = false
	client.mu.Unlock()
}

func (client *wsClient) reConnect(ctx context.Context) {
	if !client.Reconnect {
		return
	}
	for {
		time.Sleep(client.ReconnectInterval)
		if client.Connect(ctx) == nil {
			client.curReconnectNumber = 0
			if client.HandleReConn != nil {
				client.HandleReConn()
			}
			return
		}
		client.curReconnectNumber++
		if client.curReconnectNumber == client.ReconnectNumber {
			break
		}
	}
}

func (client *wsClient) Connect(ctx context.Context) error {
	conn, _, err := websocket.Dial(ctx, client.connAddr, &websocket.DialOptions{
		HTTPClient:           nil,
		HTTPHeader:           client.HTTPHeader,
		Subprotocols:         nil,
		CompressionMode:      0,
		CompressionThreshold: 0,
	})
	if err != nil {
		return err
	}
	client.mu.Lock()
	conn.SetReadLimit(1048576)
	client.conn = conn
	client.open = true
	client.mu.Unlock()

	go client.readLoop(ctx)
	return nil
}

func (client *wsClient) IsOpen() bool {
	client.mu.Lock()
	defer client.mu.Unlock()

	return client.open
}

func (client *wsClient) readLoop(ctx context.Context) {
	if !client.IsOpen() {
		return
	}
	for {
		t, payload, err := client.conn.Read(ctx)
		if err != nil {
			logger.Error("wsClient.readLoop", zap.String("connAdder", client.connAddr), zap.String("connAdder", client.proxyAddr), zap.Error(err))
			break
		}
		if t != websocket.MessageBinary {
			logger.Info("wsClient.readLoop t != websocket.BinaryMessage", zap.Int("t", int(t)))
			continue
		}
		switch payload[0] {
		case MsgTypeNotify:
			client.handleNotify(payload)
		case MsgTypeResponse:
			client.handleResponse(payload)
		default:
			logger.Info("wsClient.readLoop unknown msg type: ", zap.Uint8("value", payload[0]))
		}
		select {
		case <-ctx.Done():
			break
		default:
		}
	}
	client.Close()
	go client.reConnect(ctx)
}

func (client *wsClient) handleNotify(msg []byte) {
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[1:], wrapper)
	if err != nil {
		logger.Error("wsClient.handleNotify unmarshal error: ", zap.Error(err))
		return
	}
	pm := message.GetNotifyType(wrapper.Name)
	if pm == nil {
		logger.Error("wsClient.handleNotify unknown notify type: ", zap.String("wrapper.Name", wrapper.Name))
		return
	}
	err = proto.Unmarshal(wrapper.Data, pm)
	if err != nil {
		logger.Error("wsClient.handleNotify unmarshal error: ", zap.Error(err))
		return
	}
	client.notify <- pm
}

func (client *wsClient) handleResponse(msg []byte) {
	key := (msg[2] << 7) + msg[1]
	v, ok := client.replyMap.Load(key)
	if !ok {
		return
	}
	reply, ok := v.(*Reply)
	if !ok {
		logger.Error("wsClient.handleResponse rv not proto.Message: ", zap.Reflect("reply", reply))
		return
	}
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		logger.Error("wsClient.handleResponse unmarshal error: ", zap.Error(err))
		return
	}
	err = proto.Unmarshal(wrapper.Data, reply.out)
	if err != nil {
		logger.Error("wsClient.handleResponse unmarshal error: ", zap.Error(err))
		return
	}
	close(reply.wait)
}

func (client *wsClient) Receive() <-chan proto.Message {
	return client.notify
}

func (client *wsClient) Invoke(ctx context.Context, method string, in interface{}, out interface{}, opts ...grpc.CallOption) error {
	tokens := strings.Split(method, "/")
	api := strings.Join(tokens, ".")

	reply, err := client.SendMsg(ctx, api, in.(proto.Message))
	if err != nil {
		return err
	}
	reply.out = out.(proto.Message)

	return client.RecvMsg(ctx, reply)
}

func (client *wsClient) SendMsg(ctx context.Context, api string, in proto.Message) (_ *Reply, err error) {
	err = ErrNotConnected

	if !client.IsOpen() {
		return
	}

	var body []byte

	body, err = proto.Marshal(in)
	if err != nil {
		return
	}

	wrapper := &message.Wrapper{
		Name: api,
		Data: body,
	}

	body, err = proto.Marshal(wrapper)
	if err != nil {
		return
	}

	buff := new(bytes.Buffer)
	client.mu.Lock()
	client.msgIndex %= 255
	client.msgIndex++
	index := client.msgIndex
	client.mu.Unlock()
	buff.WriteByte(MsgTypeRequest)
	buff.WriteByte(index - (index >> 7 << 7))
	buff.WriteByte(index >> 7)
	buff.Write(body)

	err = client.conn.Write(ctx, websocket.MessageBinary, buff.Bytes())

	if err != nil {
		return
	}

	reply := &Reply{
		out:      nil,
		wait:     make(chan struct{}),
		msgIndex: index,
	}

	if _, ok := client.replyMap.LoadOrStore(reply.msgIndex, reply); ok {
		return nil, fmt.Errorf("index exists %d", reply.msgIndex)
	}

	return reply, nil
}

func (client *wsClient) RecvMsg(ctx context.Context, reply *Reply) error {
	defer client.replyMap.Delete(reply.msgIndex)
	select {
	case <-ctx.Done():
	case <-reply.wait:
	}
	return ctx.Err()
}

func (client *wsClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("implement me")
}
