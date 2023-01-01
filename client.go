package majsoul

import (
	"bytes"
	"context"
	"fmt"
	"github.com/constellation39/majsoul/logger"
	"go.uber.org/zap"
	"strings"
	"sync"

	"github.com/constellation39/majsoul/message"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

type ClientConn struct {
	ctx context.Context
	*wsClient
	mu       sync.Mutex
	msgIndex uint8
	replys   sync.Map // 回复消息 map[uint8]*Reply
	notify   chan proto.Message
}

type Reply struct {
	out  proto.Message
	wait chan struct{}
}

func NewClientConn(ctx context.Context, connAddr, proxyAddr string) (*ClientConn, error) {
	cConn := &ClientConn{
		ctx:      ctx,
		wsClient: nil,
		notify:   make(chan proto.Message, 32),
	}
	var err error

	cConn.wsClient, err = newWSClient(ctx, connAddr, proxyAddr)

	if err != nil {
		return nil, err
	}

	go cConn.loop()
	return cConn, nil
}

func (c *ClientConn) loop() {
receive:
	for {
		msg := c.wsClient.Read()
		switch msg[0] {
		case MsgTypeNotify:
			c.handleNotify(msg)
		case MsgTypeResponse:
			c.handleResponse(msg)
		default:
			logger.Info("ClientConn.loop unknown msg type: ", zap.Uint8("value", msg[0]))
		}
		select {
		case <-c.ctx.Done():
			break receive
		default:
		}
	}
}

func (c *ClientConn) handleNotify(msg []byte) {
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[1:], wrapper)
	if err != nil {
		logger.Error("ClientConn.handleNotify unmarshal error: ", zap.Error(err))
		return
	}
	pm := message.GetNotifyType(wrapper.Name)
	if pm == nil {
		logger.Error("ClientConn.handleNotify unknown notify type: ", zap.String("wrapper.Name", wrapper.Name))
		return
	}
	err = proto.Unmarshal(wrapper.Data, pm)
	if err != nil {
		logger.Error("ClientConn.handleNotify unmarshal error: ", zap.Error(err))
		return
	}
	c.notify <- pm
}

func (c *ClientConn) handleResponse(msg []byte) {
	key := (msg[2] << 7) + msg[1]
	v, ok := c.replys.Load(key)
	if !ok {
		logger.Error("ClientConn.handleResponse not found key: ", zap.Uint8("key", key))
		return
	}
	reply, ok := v.(*Reply)
	if !ok {
		logger.Error("ClientConn.handleResponse rv not proto.Message: ", zap.Reflect("reply", reply))
		return
	}
	wrapper := new(message.Wrapper)
	err := proto.Unmarshal(msg[3:], wrapper)
	if err != nil {
		logger.Error("ClientConn.handleResponse unmarshal error: ", zap.Error(err))
		return
	}
	err = proto.Unmarshal(wrapper.Data, reply.out)
	if err != nil {
		logger.Error("ClientConn.handleResponse unmarshal error: ", zap.Error(err))
		return
	}
	close(reply.wait)
}

func (c *ClientConn) Receive() <-chan proto.Message {
	return c.notify
}

func (c *ClientConn) Invoke(ctx context.Context, method string, in interface{}, out interface{}, opts ...grpc.CallOption) error {
	tokens := strings.Split(method, "/")
	api := strings.Join(tokens, ".")
	return c.Send(ctx, api, in.(proto.Message), out.(proto.Message))
}

func (c *ClientConn) Send(ctx context.Context, api string, in proto.Message, out proto.Message) error {
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

	c.mu.Lock()

	buff := new(bytes.Buffer)
	c.msgIndex %= 255
	buff.WriteByte(MsgTypeRequest)
	buff.WriteByte(c.msgIndex - (c.msgIndex >> 7 << 7))
	buff.WriteByte(c.msgIndex >> 7)
	buff.Write(body)

	err = c.wsClient.Send(buff.Bytes())
	if err != nil {
		return err
	}

	reply := &Reply{
		out:  out.(proto.Message),
		wait: make(chan struct{}),
	}
	if _, ok := c.replys.LoadOrStore(c.msgIndex, reply); ok {
		return fmt.Errorf("index exists %d", c.msgIndex)
	}
	defer c.replys.Delete(c.msgIndex)

	c.msgIndex++

	c.mu.Unlock()

	select {
	case <-reply.wait:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (c *ClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("implement me")
}
