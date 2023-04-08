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
	msgTypeNotify   uint8 = 1 // 通知
	msgTypeRequest  uint8 = 2 // 请求
	msgTypeResponse uint8 = 3 // 回复
)

type wsConfig struct {
	ConnAddress    string
	ProxyAddress   string
	RequestHeaders http.Header
}

type wsClient struct {
	*wsConfig
	conn               *websocket.Conn
	closeCh            chan struct{}
	isConnected        uint32
	messageIndex       uint32
	requestResponseMap sync.Map // map[uint8]*reply
	notify             chan proto.Message
	cancelFunc         context.CancelFunc
	errorCallback      func(error)
	reconnectHandler   func(ctx context.Context)
}

type reply struct {
	out   proto.Message
	wait  chan struct{}
	index uint8
}

func newWsClient(config *wsConfig) *wsClient {
	return &wsClient{
		wsConfig:           config,
		conn:               nil,
		closeCh:            make(chan struct{}),
		isConnected:        0,
		messageIndex:       0,
		requestResponseMap: sync.Map{},
		notify:             make(chan proto.Message, 32),
		cancelFunc:         nil,
		errorCallback:      nil,
		reconnectHandler:   nil,
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

func (client *wsClient) OnReconnect(callback func(ctx context.Context)) {
	client.reconnectHandler = callback
}

func (client *wsClient) OnError(callback func(error)) {
	client.errorCallback = callback
}

func (client *wsClient) Close() error {
	if !client.getIsConnected() {
		return fmt.Errorf("majsoul ws client is not connected")
	}
	select {
	case _, ok := <-client.closeCh:
		if !ok {
			return nil
		}
	default:
	}
	client.cancelFunc()
	close(client.closeCh)
	client.setIsConnected(false)
	if client.conn == nil {
		return fmt.Errorf("majsoul ws client connection is nil")
	}
	if err := client.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		return fmt.Errorf("majsoul ws failed to closeCh websocket connection: %w", err)
	}
	return nil
}

func (client *wsClient) reConnect(ctx context.Context) {
	for {
		select {
		case _, ok := <-client.closeCh:
			if !ok {
				return
			}
		default:
		}

		time.Sleep(time.Second)

		err := client.connect(ctx)
		if err != nil {
			if client.errorCallback != nil {
				client.errorCallback(err)
			}
			continue
		}

		if client.reconnectHandler != nil {
			client.reconnectHandler(ctx)
			return
		}
	}
}

func (client *wsClient) connect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		logger.Error("majsoul ws failed to create cookie jar: ", zap.Error(err))
		return err
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
		logger.Error("majsoul ws failed to dial: ", zap.String("connAddress", client.ConnAddress), zap.String("proxyAddress", client.ProxyAddress), zap.Error(err))
		return err
	}

	conn.SetReadLimit(1048576)
	client.conn = conn
	logger.Debug("majsoul ws connect success:", zap.String("connAddress", client.ConnAddress))
	client.setIsConnected(true)

	go client.readLoop(ctx)
	return nil
}

func (client *wsClient) Connect(ctx context.Context) error {
	ctx, client.cancelFunc = context.WithCancel(ctx)
	return client.connect(ctx)
}

func (client *wsClient) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			err := client.Close()
			if err != nil {
				logger.Error("majsoul ws closeCh: ", zap.String("connAddress", client.ConnAddress), zap.String("proxyAddress", client.ProxyAddress), zap.Error(err))
				if client.errorCallback != nil {
					client.errorCallback(err)
				}
				return
			}
			return
		default:
		}

		msgType, payload, err := client.conn.Read(ctx)
		if err != nil {
			logger.Debug("majsoul ws read: ", zap.String("connAddress", client.ConnAddress), zap.String("proxyAddress", client.ProxyAddress), zap.Error(err))
			if client.errorCallback != nil {
				client.errorCallback(err)
			}
			break
		}
		if msgType != websocket.MessageBinary {
			logger.Info("majsoul ws unsupported message types: ", zap.Int("t", int(msgType)))
			if client.errorCallback != nil {
				client.errorCallback(fmt.Errorf("majsoul ws unsupported message types: %d", msgType))
			}
			continue
		}

		if len(payload) == 0 {
			logger.Error("majsoul ws read message length is zero: ")
			if client.errorCallback != nil {
				client.errorCallback(fmt.Errorf("majsoul ws read message length is zero"))
			}
			continue
		}

		switch payload[0] {
		case msgTypeNotify:
			client.handleNotify(payload)
		case msgTypeResponse:
			client.handleResponse(payload)
		default:
			logger.Warn("majsoul ws unknown message types: ", zap.Uint8("value", payload[0]))
		}
	}

	select {
	case <-ctx.Done():
		err := client.Close()
		if err != nil {
			logger.Error("majsoul ws closeCh error: ", zap.Error(err))
		}
		return
	default:
		client.setIsConnected(false)
		client.reConnect(ctx)
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

	select {
	case client.notify <- notifyMessage:
	default:
		logger.Error("majsoul ws notify channel is full: ", zap.Reflect("wrapper", wrapper), zap.Reflect("notify message", notifyMessage))
	}

}

func (client *wsClient) handleResponse(msg []byte) {
	responseKey := (msg[2] << 7) + msg[1]

	response, ok := client.requestResponseMap.Load(responseKey)
	if !ok {
		return
	}

	r, ok := response.(*reply)
	if !ok {
		logger.Error("majsoul ws response type not proto.Message: ", zap.Reflect("reply", r))
		return
	}

	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		logger.Error("majsoul ws response message unmarshal error: ", zap.Error(err))
		return
	}

	err = proto.Unmarshal(wrapper.Data, r.out)
	if err != nil {
		logger.Error("majsoul ws response type unmarshal error: ", zap.Error(err))
		return
	}

	close(r.wait)
}

func (client *wsClient) Receive() <-chan proto.Message {
	return client.notify
}

func (client *wsClient) Invoke(ctx context.Context, method string, in interface{}, out interface{}, _ ...grpc.CallOption) error {
	tokens := strings.Split(method, "/")
	api := strings.Join(tokens, ".")

	r, err := client.sendMsg(ctx, api, in.(proto.Message))
	if err != nil {
		return err
	}
	r.out = out.(proto.Message)

	return client.recvMsg(ctx, r)
}

func (client *wsClient) sendMsg(ctx context.Context, api string, in proto.Message) (_ *reply, err error) {
	if !client.getIsConnected() {
		return nil, websocket.CloseError{Code: websocket.StatusNoStatusRcvd}
	}
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

func (client *wsClient) recvMsg(ctx context.Context, reply *reply) error {
	defer client.requestResponseMap.Delete(reply.index)
	select {
	case <-client.closeCh:
		return websocket.CloseError{Code: websocket.StatusNoStatusRcvd}
	case <-time.After(time.Minute):
		return fmt.Errorf("majsoul ws timeout waiting for response message after %s", time.Minute)
	case <-ctx.Done():
		return ctx.Err()
	case <-reply.wait:
	}
	return nil
}

func (client *wsClient) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("implement me")
}
