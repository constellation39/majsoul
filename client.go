package majsoul

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/constellation39/majsoul/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/constellation39/majsoul/message"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

var (
	ErrNotConnected   = errors.New("websocket not connected")
	ErrUrlWrongScheme = errors.New("websocket uri must start with ws or wss scheme")
)

type client struct {
	ctx context.Context

	Reconnect          bool
	ReconnectInterval  time.Duration
	ReconnectNumber    int
	curReconnectNumber int

	dialer        *websocket.Dialer
	connAddr      string
	proxyAddr     string
	requestHeader http.Header

	conn         *websocket.Conn
	httpResponse *http.Response
	isConnected  bool

	msgIndex uint8
	replyMap sync.Map // map[uint8]*Reply
	notify   chan proto.Message

	mu sync.Mutex
}

type Reply struct {
	out      proto.Message
	wait     chan struct{}
	msgIndex uint8
	timeOut  bool
}

func newClientConn(ctx context.Context, connAddr, proxyAddr string) (*client, error) {
	c := &client{
		ctx:                ctx,
		Reconnect:          false,
		ReconnectInterval:  time.Second,
		ReconnectNumber:    -1,
		curReconnectNumber: 0,
		dialer:             nil,
		connAddr:           connAddr,
		proxyAddr:          proxyAddr,
		requestHeader:      nil,
		conn:               nil,
		httpResponse:       nil,
		isConnected:        false,
		msgIndex:           0,
		replyMap:           sync.Map{},
		notify:             make(chan proto.Message, 32),
		mu:                 sync.Mutex{},
	}

	dialer, header, err := newDialer(c.connAddr, c.proxyAddr)

	if err != nil {
		return nil, err
	}

	c.requestHeader = header
	c.dialer = dialer

	c.connect()

	if err != nil {
		return nil, err
	}

	return c, nil
}

func newDialer(connAddr, proxyAddr string) (*websocket.Dialer, http.Header, error) {
	connUrl, err := url.Parse(connAddr)

	if err != nil {
		return nil, nil, err
	}

	if connUrl.Scheme != "ws" && connUrl.Scheme != "wss" {
		return nil, nil, ErrUrlWrongScheme
	}

	if err != nil {
		return nil, nil, err
	}

	dialer := &websocket.Dialer{
		NetDial:           nil,
		NetDialContext:    nil,
		NetDialTLSContext: nil,
		Proxy: func(request *http.Request) (*url.URL, error) {
			return url.Parse(proxyAddr)
		},
		TLSClientConfig:   nil,
		HandshakeTimeout:  45 * time.Second,
		ReadBufferSize:    0,
		WriteBufferSize:   0,
		WriteBufferPool:   nil,
		Subprotocols:      nil,
		EnableCompression: false,
		Jar:               nil,
	}

	dialer.Jar, _ = cookiejar.New(nil)

	header := http.Header{}
	header.Add("Accept-Encoding", "gzip, deflate, br")
	header.Add("Accept-Language", "zh-CN,zh;q=0.9,ja;q=0.8,en;q=0.7,en-GB;q=0.6,en-US;q=0.5")
	header.Add("Cache-Control", "no-cache")
	header.Add("Host", connUrl.Host)
	//header.Add("Origin", originAddr)
	header.Add("Pragma", "no-cache")
	header.Add("User-Agent", UserAgent)

	return dialer, header, nil

}

func (client *client) close() {
	client.mu.Lock()
	if client.conn != nil {
		err := client.conn.Close()
		if err != nil {
			logger.Error("client.conn.Close()", zap.Error(err))
		}
	}
	client.isConnected = false
	client.mu.Unlock()
}

func (client *client) connect() {
	for {
		conn, response, err := client.dialer.DialContext(client.ctx, client.connAddr, client.requestHeader)

		if err != nil {
			time.Sleep(client.ReconnectInterval)
			if client.ReconnectNumber == 0 || client.curReconnectNumber == client.ReconnectNumber {
				logger.Error("reConnect failed", zap.String("connAdder", client.connAddr), zap.String("connAdder", client.proxyAddr))
				break
			}
			continue
		}

		client.mu.Lock()
		client.conn = conn
		client.httpResponse = response
		client.isConnected = true
		client.mu.Unlock()
		break
	}

	go client.readLoop()
}

func (client *client) IsConnected() bool {
	client.mu.Lock()
	defer client.mu.Unlock()

	return client.isConnected
}

func (client *client) readLoop() {
	if !client.IsConnected() {
		return
	}
	for {
		t, payload, err := client.conn.ReadMessage()
		if err != nil {
			logger.Error("client.readLoop", zap.String("connAdder", client.connAddr), zap.String("connAdder", client.proxyAddr), zap.Error(err))
			break
		}
		if t != websocket.BinaryMessage {
			logger.Info("client.readLoop t != websocket.BinaryMessage", zap.Int("t", t))
			continue
		}
		switch payload[0] {
		case MsgTypeNotify:
			client.handleNotify(payload)
		case MsgTypeResponse:
			client.handleResponse(payload)
		default:
			logger.Info("client.readLoop unknown msg type: ", zap.Uint8("value", payload[0]))
		}
		select {
		case <-client.ctx.Done():
			return
		default:
		}
	}
	client.close()
	client.connect()
}

func (client *client) handleNotify(msg []byte) {
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[1:], wrapper)
	if err != nil {
		logger.Error("client.handleNotify unmarshal error: ", zap.Error(err))
		return
	}
	pm := message.GetNotifyType(wrapper.Name)
	if pm == nil {
		logger.Error("client.handleNotify unknown notify type: ", zap.String("wrapper.Name", wrapper.Name))
		return
	}
	err = proto.Unmarshal(wrapper.Data, pm)
	if err != nil {
		logger.Error("client.handleNotify unmarshal error: ", zap.Error(err))
		return
	}
	client.notify <- pm
}

func (client *client) handleResponse(msg []byte) {
	key := (msg[2] << 7) + msg[1]
	v, ok := client.replyMap.Load(key)
	if !ok {
		return
	}
	reply, ok := v.(*Reply)
	if !ok {
		logger.Error("client.handleResponse rv not proto.Message: ", zap.Reflect("reply", reply))
		return
	}
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		logger.Error("client.handleResponse unmarshal error: ", zap.Error(err))
		return
	}
	err = proto.Unmarshal(wrapper.Data, reply.out)
	if err != nil {
		logger.Error("client.handleResponse unmarshal error: ", zap.Error(err))
		return
	}
	close(reply.wait)
}

func (client *client) Receive() <-chan proto.Message {
	return client.notify
}

func (client *client) Invoke(ctx context.Context, method string, in interface{}, out interface{}, opts ...grpc.CallOption) error {
	tokens := strings.Split(method, "/")
	api := strings.Join(tokens, ".")

	reply, err := client.SendMsg(api, in.(proto.Message))
	if err != nil {
		return err
	}
	reply.out = out.(proto.Message)

	return client.RecvMsg(ctx, reply)
}

func (client *client) SendMsg(api string, in proto.Message) (_ *Reply, err error) {
	err = ErrNotConnected

	if !client.IsConnected() {
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

	client.mu.Lock()
	buff := new(bytes.Buffer)
	client.msgIndex %= 255
	buff.WriteByte(MsgTypeRequest)
	buff.WriteByte(client.msgIndex - (client.msgIndex >> 7 << 7))
	buff.WriteByte(client.msgIndex >> 7)
	buff.Write(body)

	err = client.conn.WriteMessage(websocket.BinaryMessage, buff.Bytes())

	if err != nil {
		return
	}

	reply := &Reply{
		out:      nil,
		wait:     make(chan struct{}),
		msgIndex: client.msgIndex,
	}

	client.msgIndex++
	client.mu.Unlock()

	if _, ok := client.replyMap.LoadOrStore(reply.msgIndex, reply); ok {
		return nil, fmt.Errorf("index exists %d", reply.msgIndex)
	}

	return reply, nil
}

func (client *client) RecvMsg(ctx context.Context, reply *Reply) error {
	defer client.replyMap.Delete(reply.msgIndex)

	select {
	case <-ctx.Done():
	case <-client.ctx.Done():
	case <-reply.wait:
	}
	return ctx.Err()
}

func (client *client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("implement me")
}
