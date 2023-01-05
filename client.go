package majsoul

import (
	"bytes"
	"context"
	"fmt"
	"github.com/constellation39/majsoul/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io"
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

type Client struct {
	ctx       context.Context
	connAddr  string
	proxyAddr string
	conn      *websocket.Conn
	mu        sync.Mutex
	msgIndex  uint8
	replyMap  sync.Map // 回复消息 map[uint8]*Reply
	notify    chan proto.Message
	reConn    chan struct{}
}

type Reply struct {
	out  proto.Message
	wait chan struct{}
}

func NewClientConn(ctx context.Context, connAddr, proxyAddr string) (*Client, error) {
	client := &Client{
		ctx:       ctx,
		connAddr:  connAddr,
		proxyAddr: proxyAddr,
		conn:      nil,
		mu:        sync.Mutex{},
		msgIndex:  0,
		replyMap:  sync.Map{},
		notify:    make(chan proto.Message, 32),
	}
	var err error

	client.conn, err = newConn(ctx, connAddr, proxyAddr)

	if err != nil {
		return nil, err
	}

	go client.readLoop()
	return client, nil
}

func newConn(ctx context.Context, connAddr, proxyAddr string) (*websocket.Conn, error) {
	connUrl, err := url.Parse(connAddr)

	if err != nil {
		return nil, err
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

	conn, response, err := dialer.DialContext(ctx, connAddr, header)

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	return conn, nil
}

func (client *Client) readLoop() {
	var t int
	var payload []byte
	var err error
	for {
		t, payload, err = client.conn.ReadMessage()
		if err != nil {
			logger.Error("Client.readLoop", zap.Error(err))
			break
		}
		if t != websocket.BinaryMessage {
			logger.Info("Client.readLoop t != websocket.BinaryMessage", zap.Int("t", t))
			continue
		}
		switch payload[0] {
		case MsgTypeNotify:
			client.handleNotify(payload)
		case MsgTypeResponse:
			client.handleResponse(payload)
		default:
			logger.Info("Client.readLoop unknown msg type: ", zap.Uint8("value", payload[0]))
		}
		select {
		case <-client.ctx.Done():
			return
		default:
		}
	}
	err = client.conn.Close()
	if err != nil {
		logger.Error("client.conn.Close()", zap.Error(err))
		return
	}
}

func (client *Client) handleNotify(msg []byte) {
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[1:], wrapper)
	if err != nil {
		logger.Error("Client.handleNotify unmarshal error: ", zap.Error(err))
		return
	}
	pm := message.GetNotifyType(wrapper.Name)
	if pm == nil {
		logger.Error("Client.handleNotify unknown notify type: ", zap.String("wrapper.Name", wrapper.Name))
		return
	}
	err = proto.Unmarshal(wrapper.Data, pm)
	if err != nil {
		logger.Error("Client.handleNotify unmarshal error: ", zap.Error(err))
		return
	}
	client.notify <- pm
}

func (client *Client) handleResponse(msg []byte) {
	key := (msg[2] << 7) + msg[1]
	v, ok := client.replyMap.Load(key)
	if !ok {
		logger.Error("Client.handleResponse not found key: ", zap.Uint8("key", key))
		return
	}
	reply, ok := v.(*Reply)
	if !ok {
		logger.Error("Client.handleResponse rv not proto.Message: ", zap.Reflect("reply", reply))
		return
	}
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		logger.Error("Client.handleResponse unmarshal error: ", zap.Error(err))
		return
	}
	err = proto.Unmarshal(wrapper.Data, reply.out)
	if err != nil {
		logger.Error("Client.handleResponse unmarshal error: ", zap.Error(err))
		return
	}
	close(reply.wait)
}

func (client *Client) Receive() <-chan proto.Message {
	return client.notify
}

func (client *Client) Invoke(ctx context.Context, method string, in interface{}, out interface{}, opts ...grpc.CallOption) error {
	tokens := strings.Split(method, "/")
	api := strings.Join(tokens, ".")
	return client.Send(ctx, api, in.(proto.Message), out.(proto.Message))
}

func (client *Client) Send(ctx context.Context, api string, in proto.Message, out proto.Message) error {
	body, err := proto.Marshal(in.(proto.Message))
	if err != nil {
		return err
	}

	wrapper := &message.Wrapper{
		Name: api,
		Data: body,
	}

	body, err = proto.Marshal(wrapper)
	if err != nil {
		return err
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
		return err
	}

	reply := &Reply{
		out:  out.(proto.Message),
		wait: make(chan struct{}),
	}
	if _, ok := client.replyMap.LoadOrStore(client.msgIndex, reply); ok {
		return fmt.Errorf("index exists %d", client.msgIndex)
	}
	defer client.replyMap.Delete(client.msgIndex)

	client.msgIndex++

	client.mu.Unlock()

	select {
	case <-reply.wait:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (client *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("implement me")
}
