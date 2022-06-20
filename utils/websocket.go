package utils

import (
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"sync"
)

type WSClient struct {
	connAddr string // 连接地址
	mu       *sync.Mutex
	header   http.Header
	conn     *websocket.Conn // websocket连接
}

func NewWSClient(addr string) *WSClient {
	return &WSClient{
		connAddr: addr,
		mu:       &sync.Mutex{},
	}
}

func (client *WSClient) Connect() error {
	conn, response, err := websocket.DefaultDialer.Dial(client.connAddr, client.header)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	client.conn = conn
	return nil
}

func (client *WSClient) Read() []byte {
	t, payload, err := client.conn.ReadMessage()
	if err != nil {
		return []byte{}
	}

	if t != websocket.BinaryMessage {
		return []byte{}
	}
	return payload
}

func (client *WSClient) Send(body []byte) error {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.conn.WriteMessage(websocket.BinaryMessage, body)
}

func (client *WSClient) Close() error {
	if client.conn != nil {
		return client.conn.Close()
	}
	return errors.New("websocket connection is nil")
}
