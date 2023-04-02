package majsoul

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/constellation39/majsoul/logger"
	"go.uber.org/zap"
	"nhooyr.io/websocket"

	"github.com/constellation39/majsoul/message"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	MsgTypeNotify   uint8 = 1 // 通知
	MsgTypeRequest  uint8 = 2 // 请求
	MsgTypeResponse uint8 = 3 // 回复
)

type wsConfig struct {
	ConnAddress    string
	ProxyAddress   string
	RequestHeaders http.Header
}

type wsClient struct {
	*wsConfig
	conn               *websocket.Conn
	close              chan struct{}
	isConnected        uint32
	messageIndex       uint32
	requestResponseMap sync.Map // map[uint8]*Reply
	notify             chan proto.Message

	closeHandler     func()
	reconnectHandler func(ctx context.Context)
}

type Reply struct {
	out   proto.Message
	wait  chan struct{}
	index uint8
}

func newWsClient(config *wsConfig) *wsClient {
	return &wsClient{
		wsConfig:           config,
		conn:               nil,
		close:              make(chan struct{}),
		isConnected:        0,
		messageIndex:       0,
		requestResponseMap: sync.Map{},
		notify:             make(chan proto.Message, 32),
	}
}

func (client *wsClient) setIsConnected(connected bool) {
	var newVal uint32
	if connected {
		newVal = 1
	}
	atomic.StoreUint32(&client.isConnected, newVal)
}

func (client *wsClient) getIsConnected() bool {
	return atomic.LoadUint32(&client.isConnected) == 1
}

func (client *wsClient) OnReconnect(callbrak func(ctx context.Context)) {
	client.reconnectHandler = callbrak
}

func (client *wsClient) OnClose(callbrak func()) {
	client.closeHandler = callbrak
}

func (client *wsClient) Close() {
	if client.getIsConnected() {
		close(client.close)
		client.setIsConnected(false)
		if client.conn != nil {
			if err := client.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
				logger.Error("majsoul ws failed to close websocket connection: ", zap.Error(err))
			}
		}
		if client.closeHandler != nil {
			client.closeHandler()
		}
	}
}

func (client *wsClient) reConnect(ctx context.Context) {
	for {
		select {
		case _, ok := <-client.close:
			if !ok {
				return
			}
		default:
		}

		time.Sleep(time.Second)

		err := client.Connect(ctx)
		if err != nil {
			continue
		}

		if client.reconnectHandler != nil {
			client.reconnectHandler(ctx)
			return
		}
	}
}

func (client *wsClient) Connect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		logger.Error("majsoul ws failed to create cookie jar: ", zap.Error(err))
	}

	httpClient := &http.Client{
		Transport:     &http.Transport{},
		CheckRedirect: nil,
		Jar:           jar,
		Timeout:       time.Minute,
	}

	if len(client.wsConfig.ProxyAddress) > 0 {
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(client.wsConfig.ProxyAddress)
		}
		transport := &http.Transport{Proxy: proxy}
		httpClient.Transport = transport
	}

	conn, _, err := websocket.Dial(ctx, client.ConnAddress, &websocket.DialOptions{
		HTTPClient:           httpClient,
		HTTPHeader:           client.RequestHeaders,
		Subprotocols:         nil,
		CompressionMode:      0,
		CompressionThreshold: 0,
	})
	if err != nil {
		return err
	}

	conn.SetReadLimit(1048576)
	client.conn = conn
	logger.Debug("majsoul ws connect success:", zap.String("connAddress", client.ConnAddress))
	client.setIsConnected(true)

	go client.readLoop(ctx)
	return nil
}

func (client *wsClient) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			client.Close()
			return
		default:
		}

		msgType, payload, err := client.conn.Read(ctx)
		if err != nil {
			logger.Debug("majsoul ws read: ", zap.String("connAddress", client.ConnAddress), zap.String("proxyAddress", client.ProxyAddress), zap.Error(err))
			break
		}
		if msgType != websocket.MessageBinary {
			logger.Info("majsoul ws unsupported message types: ", zap.Int("t", int(msgType)))
			continue
		}

		if len(payload) == 0 {
			logger.Error("majsoul ws read message length is zero: ")
			continue
		}

		switch payload[0] {
		case MsgTypeNotify:
			client.handleNotify(payload)
		case MsgTypeResponse:
			client.handleResponse(payload)
		default:
			logger.Info("majsoul ws unknown message types: ", zap.Uint8("value", payload[0]))
		}
	}

	select {
	case <-ctx.Done():
		client.Close()
		return
	default:
		client.setIsConnected(false)
		go client.reConnect(ctx)
	}
}

func (client *wsClient) handleNotify(msg []byte) {
	wrapper := new(message.Wrapper)

	err := proto.Unmarshal(msg[1:], wrapper)
	if err != nil {
		logger.Error("majsoul ws notify messages unmarshal error: ", zap.Error(err))
		return
	}

	notifyMessage := message.GetNotifyType(wrapper.Name)
	if notifyMessage == nil {
		logger.Error("majsoul ws unknown notify type: ", zap.String("name", wrapper.Name))
		return
	}

	err = proto.Unmarshal(wrapper.Data, notifyMessage)
	if err != nil {
		logger.Error("majsoul ws notify type unmarshal error: ", zap.Reflect("notify type", notifyMessage), zap.Error(err))
		return
	}

	client.notify <- notifyMessage
}

func (client *wsClient) handleResponse(msg []byte) {
	responseKey := uint8((msg[2] << 7) + msg[1])

	response, ok := client.requestResponseMap.Load(responseKey)
	if !ok {
		return
	}

	reply, ok := response.(*Reply)
	if !ok {
		logger.Error("ws response type not proto.Message: ", zap.Reflect("reply", reply))
		return
	}

	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		logger.Error("ws response message unmarshal error: ", zap.Error(err))
		return
	}

	err = proto.Unmarshal(wrapper.Data, reply.out)
	if err != nil {
		logger.Error("ws response type unmarshal error: ", zap.Error(err))
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
	if !client.getIsConnected() {
		return nil, websocket.CloseError{}
	}
	var body []byte

	body, err = proto.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("ws failed to marshal input message: %v, error: %w", in, err)
	}

	wrapper := &message.Wrapper{
		Name: api,
		Data: body,
	}

	body, err = proto.Marshal(wrapper)
	if err != nil {
		return nil, fmt.Errorf("ws failed to marshal input message: %w", err)
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
	buff.WriteByte(MsgTypeRequest)
	buff.WriteByte(indexUint8 - (indexUint8 >> 7 << 7))
	buff.WriteByte(indexUint8 >> 7)
	buff.Write(body)

	err = client.conn.Write(ctx, websocket.MessageBinary, buff.Bytes())

	if err != nil {
		return
	}

	reply := &Reply{
		out:   nil,
		wait:  make(chan struct{}),
		index: indexUint8,
	}

	if _, ok := client.requestResponseMap.LoadOrStore(reply.index, reply); ok {
		return nil, fmt.Errorf("ws message index %d already exists in the requestResponseMap", reply.index)
	}

	return reply, nil
}

func (client *wsClient) RecvMsg(ctx context.Context, reply *Reply) error {
	defer client.requestResponseMap.Delete(reply.index)
	select {
	case <-client.close:
		return websocket.CloseError{Code: websocket.StatusNoStatusRcvd}
	case <-time.After(time.Minute):
		return fmt.Errorf("ws timeout waiting for response message after %s", time.Minute)
	case <-ctx.Done():
	case <-reply.wait:
	}
	return ctx.Err()
}

func (client *wsClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("implement me")
}
