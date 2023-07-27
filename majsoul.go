// Package majsoul provides an interface to interact with the Majsoul game server.
package majsoul

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"github.com/constellation39/majsoul/network"
	"github.com/constellation39/majsoul/utils"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"nhooyr.io/websocket"
	"reflect"
	"strings"
	"time"
)

// ServerAddress represents the server configuration for Majsoul.
type ServerAddress struct {
	ServerAddress  string `json:"serverAddress"`
	GatewayAddress string `json:"gatewayAddress"`
	GameAddress    string `json:"gameAddress"`
}

// ServerAddressList is the pre-set list of server addresses.
var ServerAddressList = []*ServerAddress{
	{
		ServerAddress:  "https://game.maj-soul.net",
		GatewayAddress: "wss://gateway-v2.maj-soul.com/gateway",
		GameAddress:    "wss://gateway-v2.maj-soul.com/game-gateway",
	},
	{
		ServerAddress:  "https://game.maj-soul.com",
		GatewayAddress: "wss://gateway-v2.maj-soul.com/gateway",
		GameAddress:    "wss://gateway-v2.maj-soul.com/game-gateway",
	},
}

// Config the configuration for Majsoul.
type Config struct {
	ProxyAddress string
}

// MajSoul represents the main class for interacting with the Majsoul game server.
type MajSoul struct {
	config  *Config          // Config passed to the class
	Request *network.Request // HTTP request sent to Majsoul
	Version *Version         // Current version number

	message.LobbyClient                      // LobbyClient is the interface for interacting with the Majsoul lobby
	message.FastTestClient                   // FastTestClient is the interface for interacting with the Majsoul game table
	lobbyClientConn        *network.WsClient // Connection used by LobbyClient
	fastTestClientConn     *network.WsClient // Connection used by FastTestClient
	ServerAddress          *ServerAddress    // Server address being used
	UUID                   string            // UUID

	handleMap                  map[string]*subscribe // Map of registered addresses
	onGatewayReconnectCallBack func()                // Callback for gateway server reconnection
	onGameReconnectCallBack    func()                // Callback for game server reconnection
}

// Subscribe subscribed message.
type subscribe struct {
	in   reflect.Type
	call reflect.Value
}

// NewMajSoul creates a new instance of MajSoul with the given configuration.
func NewMajSoul(config *Config) *MajSoul {
	majSoul := &MajSoul{
		config:                     config,
		Request:                    nil,
		Version:                    nil,
		LobbyClient:                nil,
		FastTestClient:             nil,
		lobbyClientConn:            nil,
		fastTestClientConn:         nil,
		ServerAddress:              nil,
		UUID:                       utils.UUID(),
		handleMap:                  make(map[string]*subscribe),
		onGatewayReconnectCallBack: nil,
		onGameReconnectCallBack:    nil,
	}
	// 特殊注册的消息
	majSoul.Handle(majSoul.ActionPrototype)
	return majSoul
}

// Handle registers callbacks for handling specific actions.
// The callback should be a function with the following signature: func(*MajSoul, proto.Message).
func (majSoul *MajSoul) Handle(callbacks ...interface{}) {
	majSoul.register(&majSoul.handleMap, callbacks...)
}

func (majSoul *MajSoul) register(mapSubscribe *map[string]*subscribe, callbacks ...interface{}) {
	for index, callback := range callbacks {
		valueOf := reflect.ValueOf(callback)
		if valueOf.IsNil() {
			panic(fmt.Sprintf("index.%d callback is nil", index))
		}
		if valueOf.Type().NumIn() != 2 {
			panic(fmt.Sprintf("index.%d callback input parameter length not 2 want (*MajSoul,proto.message as majsoul) length = %d", index, valueOf.Type().NumIn()))
		}
		inType0 := valueOf.Type().In(0)
		if inType0.Kind() != reflect.Pointer {
			panic(fmt.Sprintf("index.%d callback input parameter type not Pointer kind = %d", index, inType0.Kind()))
		}
		if inType0.Elem().Name() != "MajSoul" {
			panic(fmt.Sprintf("index.%d callback input parameter 0 type not *MajSoul name = %s", index, inType0.Elem().Name()))
		}
		inType1 := valueOf.Type().In(1)
		if inType1.Kind() != reflect.Pointer {
			panic(fmt.Sprintf("index.%d callback input parameter type not Pointer kind = %d", index, inType1.Kind()))
		}
		name := inType1.Elem().Name()
		(*mapSubscribe)[name] = &subscribe{
			in:   inType1.Elem(),
			call: valueOf,
		}
	}
}

// LookupGateway looks up the gateway server and establishes a connection.
func (majSoul *MajSoul) LookupGateway(ctx context.Context, serverAddressList []*ServerAddress) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, serverAddress := range serverAddressList {
		var connUrl *url.URL
		connUrl, err = url.Parse(serverAddress.ServerAddress)

		if err != nil {
			return fmt.Errorf("parse url error %v", err)
		}

		header := http.Header{}
		header.Add("Accept-Encoding", "gzip, deflate, br")
		header.Add("Accept-Language", "zh-CN,zh;q=0.9,ja;q=0.8,en;q=0.7,en-GB;q=0.6,en-US;q=0.5")
		header.Add("Cache-Control", "no-cache")
		header.Add("Host", connUrl.Host)
		header.Add("Origin", serverAddress.ServerAddress)
		header.Add("Pragma", "no-cache")
		header.Add("User-Agent", network.UserAgent)

		{ // HTTP
			httpClient := http.Client{
				Jar: func() http.CookieJar {
					jar, err := cookiejar.New(nil)
					if err != nil {
						panic(err)
					}
					return jar
				}(),
				Transport: func() http.RoundTripper {
					if len(majSoul.config.ProxyAddress) == 0 {
						return nil
					}
					proxy := func(_ *http.Request) (*url.URL, error) {
						return url.Parse(majSoul.config.ProxyAddress)
					}
					return &http.Transport{Proxy: proxy}
				}(),
				Timeout: time.Second * 5,
			}
			majSoul.Request = network.NewRequest(serverAddress.ServerAddress, header, httpClient)
			err = majSoul.getVersion()
			if err != nil {
				continue
			}
		}

		{ // Websocket
			httpClient := http.Client{
				Jar: func() http.CookieJar {
					jar, err := cookiejar.New(nil)
					if err != nil {
						panic(err)
					}
					return jar
				}(),
				Transport: func() http.RoundTripper {
					if len(majSoul.config.ProxyAddress) == 0 {
						return nil
					}
					proxy := func(_ *http.Request) (*url.URL, error) {
						return url.Parse(majSoul.config.ProxyAddress)
					}
					return &http.Transport{Proxy: proxy}
				}(),
				Timeout: time.Second * 5,
			}
			majSoul.lobbyClientConn = network.NewWsClient(serverAddress.GatewayAddress, websocket.DialOptions{
				HTTPClient:           &httpClient,
				HTTPHeader:           header,
				Subprotocols:         nil,
				CompressionMode:      0,
				CompressionThreshold: 0,
			})
			err = majSoul.lobbyClientConn.Connect(ctx)
			if err != nil {
				continue
			}
			majSoul.ServerAddress = serverAddress
			majSoul.LobbyClient = message.NewLobbyClient(majSoul.lobbyClientConn)
		}

		go majSoul.heatbeat()
		go majSoul.readLobbyClientConn()

		return nil
	}

	majSoul.Request = nil
	return fmt.Errorf("no server")
}

func (majSoul *MajSoul) heatbeat() {
	// Gateway 心跳包 5 秒一次
	t5 := time.NewTicker(time.Second * 5)
	// Game 心跳包 2 秒一次
	t2 := time.NewTicker(time.Second * 2)
	var err error
	for {
		select {
		case <-t5.C:
			if majSoul.fastTestClientConn != nil {
				continue
			}
			if majSoul.lobbyClientConn == nil {
				continue
			}
			{
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				_, err = majSoul.Heatbeat(ctx, &message.ReqHeatBeat{})
				if err != nil {
					logger.Error("gateway heatbeat error", zap.Error(err))
				}
			}
		case <-t2.C:
			if majSoul.fastTestClientConn == nil {
				continue
			}
			{
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				_, err = majSoul.CheckNetworkDelay(ctx, &message.ReqCommon{})
				if err != nil {
					logger.Error("game checkNetworkDelay error", zap.Error(err))
				}
			}
		}
	}
}

// Version represents the version information for the client.
type Version struct {
	Version      string `json:"version"`
	ForceVersion string `json:"force_version"`
	Code         string `json:"code"`
}

// Web return version web format
// field Version "0.10.113.web"
// return "web-0.10.113"
func (v *Version) Web() string {
	return fmt.Sprintf("web-%s", v.Version[:len(v.Version)-2])
}

func (majSoul *MajSoul) version() (*Version, error) {
	r := int(rand.Float32()*1e9) + int(rand.Float32()*1e9)
	body, err := majSoul.Request.Get(fmt.Sprintf("1/version.json?randv=%d", r))
	if err != nil {
		return nil, err
	}
	version := new(Version)
	err = json.Unmarshal(body, version)
	if err != nil {
		return nil, err
	}
	return version, nil
}

func (majSoul *MajSoul) getVersion() (err error) {
	majSoul.Version, err = majSoul.version()
	return
}

// ConnGame connects to the game server.
func (majSoul *MajSoul) ConnGame(ctx context.Context) (err error) {
	var connUrl *url.URL
	connUrl, err = url.Parse(majSoul.ServerAddress.GameAddress)

	if err != nil {
		return fmt.Errorf("parse url error %v", err)
	}

	header := http.Header{}
	header.Add("Accept-Encoding", "gzip, deflate, br")
	header.Add("Accept-Language", "zh-CN,zh;q=0.9,ja;q=0.8,en;q=0.7,en-GB;q=0.6,en-US;q=0.5")
	header.Add("Cache-Control", "no-cache")
	header.Add("Host", connUrl.Host)
	header.Add("Origin", majSoul.ServerAddress.GameAddress)
	header.Add("Pragma", "no-cache")
	header.Add("User-Agent", network.UserAgent)
	httpClient := http.Client{
		Jar: func() http.CookieJar {
			jar, err := cookiejar.New(nil)
			if err != nil {
				panic(err)
			}
			return jar
		}(),
		Transport: func() http.RoundTripper {
			if len(majSoul.config.ProxyAddress) == 0 {
				return nil
			}
			proxy := func(_ *http.Request) (*url.URL, error) {
				return url.Parse(majSoul.config.ProxyAddress)
			}
			return &http.Transport{Proxy: proxy}
		}(),
		Timeout: time.Second * 5,
	}

	majSoul.fastTestClientConn = network.NewWsClient(majSoul.ServerAddress.GameAddress, websocket.DialOptions{
		HTTPClient:           &httpClient,
		HTTPHeader:           header,
		Subprotocols:         nil,
		CompressionMode:      0,
		CompressionThreshold: 0,
	})
	err = majSoul.fastTestClientConn.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connect game server failed error %v", err)
	}
	majSoul.FastTestClient = message.NewFastTestClient(majSoul.fastTestClientConn)
	go majSoul.readFastTestClientConn()
	return nil
}

func (majSoul *MajSoul) readLobbyClientConn() {
	if majSoul.lobbyClientConn == nil {
		panic("lobbyClient Conn is nil")
	}
	majSoul.lobbyClientConn.ReconnectHandler = majSoul.onGatewayReconnectCallBack
	receive := majSoul.lobbyClientConn.Receive()
	for wrapper := range receive {
		majSoul.callHandleMap(wrapper)
	}
}

func (majSoul *MajSoul) readFastTestClientConn() {
	if majSoul.fastTestClientConn == nil {
		panic("fastTestClient Conn is nil")
	}
	majSoul.fastTestClientConn.ReconnectHandler = majSoul.onGameReconnectCallBack
	receive := majSoul.fastTestClientConn.Receive()
	for wrapper := range receive {
		majSoul.callHandleMap(wrapper)
	}
}

func (majSoul *MajSoul) callHandleMap(wrapper *message.Wrapper) {
	token := strings.Split(wrapper.Name, ".")
	if len(token) == 0 {
		panic(fmt.Sprintf("wrapper.Name == %s", wrapper.Name))
	}
	name := token[len(token)-1]
	ss, ok := majSoul.handleMap[name]
	if !ok {
		logger.Info("unregistered notify", zap.String("name", wrapper.Name))
		return
	}
	inValue := reflect.New(ss.in)
	notify, ok := inValue.Interface().(proto.Message)
	if !ok {
		panic("in type not implements proto.Message")
	}
	err := proto.Unmarshal(wrapper.Data, notify)
	if err != nil {
		panic(fmt.Sprintf("proto unmarshal error %v", err))
	}
	ss.call.Call([]reflect.Value{reflect.ValueOf(majSoul), inValue})
}

// OnGatewayReconnect sets the callback for when the connection to the gateway server is reestablished.
func (majSoul *MajSoul) OnGatewayReconnect(callback func()) {
	majSoul.onGatewayReconnectCallBack = callback
	if majSoul.lobbyClientConn != nil {
		majSoul.lobbyClientConn.ReconnectHandler = callback
	}
}

// OnGameReconnect sets the callback for when the connection to the game server is reestablished.
func (majSoul *MajSoul) OnGameReconnect(callback func()) {
	majSoul.onGameReconnectCallBack = callback
	if majSoul.fastTestClientConn != nil {
		majSoul.fastTestClientConn.ReconnectHandler = callback
	}
}

// Login logs in to the Majsoul server with the given account and password.
func (majSoul *MajSoul) Login(ctx context.Context, account, password string) (*message.ResLogin, error) {
	if len(account) == 0 {
		return nil, fmt.Errorf("account is null")
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("password is null")
	}
	var t uint32
	if !strings.Contains(account, "@") {
		t = 1
	}
	reqLogin := &message.ReqLogin{
		Account:   account,
		Password:  utils.HashPassword(password),
		Reconnect: false,
		Device: &message.ClientDeviceInfo{
			Platform:       "pc",
			Hardware:       "pc",
			Os:             "windows",
			OsVersion:      "win10",
			IsBrowser:      true,
			Software:       "Chrome",
			SalePlatform:   "web",
			HardwareVendor: "",
			ModelNumber:    "",
			ScreenWidth:    uint32(rand.Int31n(400) + 914),
			ScreenHeight:   uint32(rand.Int31n(200) + 1316),
		},
		RandomKey: majSoul.UUID,
		ClientVersion: &message.ClientVersionInfo{
			Resource: majSoul.Version.Version,
			Package:  "",
		},
		GenAccessToken:    true,
		CurrencyPlatforms: []uint32{2, 6, 8, 10, 11},
		// 电话1 邮箱0
		Type:                t,
		Version:             0,
		ClientVersionString: majSoul.Version.Web(),
	}
	resLogin, err := majSoul.LobbyClient.Login(ctx, reqLogin)
	if err != nil {
		return nil, err
	}
	return resLogin, nil
}

// ActionPrototype handles actions from the server.
func (majSoul *MajSoul) ActionPrototype(_ *MajSoul, actionPrototype *message.ActionPrototype) {
	utils.DecodeActionPrototype(actionPrototype)
	ss, ok := majSoul.handleMap[actionPrototype.Name]
	if !ok {
		logger.Debug("unregistered action", zap.String("name", actionPrototype.Name))
		return
	}
	inValue := reflect.New(ss.in)
	notify, ok := inValue.Interface().(proto.Message)
	if !ok {
		panic("in type not implements proto.Message")
	}
	err := proto.Unmarshal(actionPrototype.Data, notify)
	if err != nil {
		panic(fmt.Sprintf("proto unmarshal error %v", err))
	}
	ss.call.Call([]reflect.Value{reflect.ValueOf(majSoul), inValue})
}
