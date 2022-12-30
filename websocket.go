package majsoul

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type wsClient struct {
	conn *websocket.Conn // websocket连接
}

func newWSClient(ctx context.Context, connAddr, proxyAddr string) (*wsClient, error) {
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

	return &wsClient{
		conn: conn,
	}, nil
}

func (client *wsClient) Read() []byte {
	t, payload, err := client.conn.ReadMessage()
	if err != nil {
		return []byte{}
	}

	if t != websocket.BinaryMessage {
		return []byte{}
	}
	return payload
}

func (client *wsClient) Send(body []byte) error {
	return client.conn.WriteMessage(websocket.BinaryMessage, body)
}

func (client *wsClient) Close() error {
	if client.conn != nil {
		return client.conn.Close()
	}
	return errors.New("websocket connection is nil")
}
