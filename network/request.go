package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const (
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54"
)

// Request is a struct that helps in sending network requests.
type Request struct {
	Host    string
	header  http.Header
	client  http.Client
	rwMutex sync.RWMutex
}

// NewRequest creates a new Request object.
func NewRequest(hostAddress string, header http.Header, client http.Client) *Request {
	r := &Request{
		Host:   hostAddress,
		header: header,
		client: client,
	}
	return r
}

// Get sends a GET request to the specified path.
func (request *Request) Get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", request.Host, path), nil)
	if err != nil {
		return nil, err
	}

	return request.do(req)
}

// Post sends a POST request to the specified path with the provided body.
func (request *Request) Post(path string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", request.Host, path), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return request.do(req)
}

// do sends the request and returns the response.
func (request *Request) do(req *http.Request) ([]byte, error) {
	req.Header = request.header
	res, err := request.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode: %s", res.Status)
	}

	defer res.Body.Close() // Just close the body, ignore error if any.

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resData, nil
}

// GetHeader returns the value(s) of the header for the provided key.
func (request *Request) GetHeader(key string) ([]string, bool) {
	request.rwMutex.RLock()
	defer request.rwMutex.RUnlock()

	if kv, ok := request.header[key]; ok {
		return kv, ok
	}
	return nil, false
}

// DelHeader removes the header for the provided key.
func (request *Request) DelHeader(key string) *Request {
	request.rwMutex.Lock()
	defer request.rwMutex.Unlock()

	request.header.Del(key)

	return request
}

// SetHeader sets the value of the header for the provided key.
func (request *Request) SetHeader(key, value string) *Request {
	request.rwMutex.Lock()
	defer request.rwMutex.Unlock()

	request.header.Set(key, value)

	return request
}

// AddHeader adds a new header value for the provided key.
func (request *Request) AddHeader(key, value string) *Request {
	request.rwMutex.Lock()
	defer request.rwMutex.Unlock()

	request.header.Add(key, value)

	return request
}
