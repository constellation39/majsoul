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

type Request struct {
	Host    string
	header  http.Header
	client  http.Client
	rwMutex sync.RWMutex
}

func NewRequest(hostAddress string, client http.Client) *Request {
	r := &Request{
		Host:   hostAddress,
		header: http.Header{},
		client: client,
	}
	//jar, _ := cookiejar.New(nil)
	//r := &Request{
	//	Host:   hostAddr,
	//	header: http.Header{},
	//	client: http.Client{
	//		Jar:     jar,
	//		Timeout: time.Second * 5,
	//	},
	//}
	//if len(proxyAddr) > 0 {
	//	proxy := func(_ *http.Request) (*url.URL, error) {
	//		return url.Parse(proxyAddr)
	//	}
	//	transport := &http.Transport{Proxy: proxy}
	//	r.client.Transport = transport
	//}
	r.AddHeader("user-agent", UserAgent)
	r.AddHeader("accept", "application/json, text/plain, */*")
	r.AddHeader("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	r.AddHeader("cache-control", "no-cache")
	r.AddHeader("content-type", "application/json;charset=UTF-8")
	r.AddHeader("dnt", "1")
	return r
}

func (request *Request) Get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", request.Host, path), nil)
	if err != nil {
		return nil, err
	}

	return request.do(req)
}

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

func (request *Request) do(req *http.Request) ([]byte, error) {
	req.Header = request.header
	res, err := request.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%s", res.Status)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resData, err
}

func (request *Request) GetHeader(key string) ([]string, bool) {
	request.rwMutex.Lock()
	defer request.rwMutex.Unlock()

	if kv, ok := request.header[key]; ok {
		return kv, ok
	}
	return nil, false
}

func (request *Request) DelHeader(key string) *Request {
	request.rwMutex.Lock()
	defer request.rwMutex.Unlock()

	request.header.Del(key)

	return request
}

func (request *Request) SetHeader(key, value string) *Request {
	request.rwMutex.Lock()
	defer request.rwMutex.Unlock()

	request.header.Set(key, value)

	return request
}

func (request *Request) AddHeader(key, value string) *Request {
	request.rwMutex.Lock()
	defer request.rwMutex.Unlock()

	request.header.Add(key, value)

	return request
}
