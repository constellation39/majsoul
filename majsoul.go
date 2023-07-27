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

// ServerAddress majsoul server config
type ServerAddress struct {
	ServerAddress  string `json:"serverAddress"`
	GatewayAddress string `json:"gatewayAddress"`
	GameAddress    string `json:"gameAddress"`
}

var ServerAddressList = []*ServerAddress{
	{
		ServerAddress:  "https://game.maj-soul.net",
		GatewayAddress: "wss://gateway-hw.maj-soul.com/gateway",
		GameAddress:    "wss://gateway-hw.maj-soul.com/game-gateway",
	},
	{
		ServerAddress:  "https://game.maj-soul.com",
		GatewayAddress: "wss://gateway-v2.maj-soul.com/gateway",
		GameAddress:    "wss://gateway-v2.maj-soul.com/game-gateway",
	},
}

type Config struct {
	ProxyAddress string
}

type MajSoul struct {
	config  *Config
	Request *network.Request // 对于雀魂发送的 http 请求
	Version *Version

	message.LobbyClient                      // message.LobbyClient 更多时候在大厅时调用的是该接口
	message.FastTestClient                   // message.FastTestClient 场景处于游戏桌面时调用该接口
	lobbyClientConn        *network.WsClient // lobbyConn 是 message.LobbyClient 使用的连接
	fastTestClientConn     *network.WsClient // fastTestConn 是 message.FastTestClient 使用的连接
	ServerAddress          *ServerAddress
	Account                *message.Account     // 该字段应在登录成功后访问
	GameInfo               *message.ResAuthGame // 该字段应在进入游戏桌面后访问
	AccessToken            string               // 验证身份时使用 的 token
	ConnectToken           string               // 重连时使用的 token
	GameUuid               string               // 是否在游戏中
	UUID                   string               // UUID

	handleMap                  map[string]*subscribe
	onGatewayReconnectCallBack func(*message.ResLogin)
	onGameReconnectCallBack    func(*message.ResSyncGame)
}

type subscribe struct {
	in   reflect.Type
	call reflect.Value
}

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
		Account:                    nil,
		GameInfo:                   nil,
		AccessToken:                "",
		ConnectToken:               "",
		GameUuid:                   "",
		UUID:                       utils.UUID(),
		handleMap:                  make(map[string]*subscribe),
		onGatewayReconnectCallBack: nil,
		onGameReconnectCallBack:    nil,
	}

	majSoul.Handle(majSoul.NotifyRoomGameStart, majSoul.ActionPrototype)
	return majSoul
}

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
			majSoul.Request = network.NewRequest(serverAddress.ServerAddress, httpClient)
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
				HTTPHeader:           nil,
				Subprotocols:         nil,
				CompressionMode:      0,
				CompressionThreshold: 0,
			}, majSoul.onGatewayReconnect)
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
	return nil
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

func (majSoul *MajSoul) ConnGame(ctx context.Context) (err error) {
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
		HTTPHeader:           nil,
		Subprotocols:         nil,
		CompressionMode:      0,
		CompressionThreshold: 0,
	}, majSoul.onGatewayReconnect)
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
	receive := majSoul.lobbyClientConn.Receive()
	for wrapper := range receive {
		majSoul.callHandleMap(wrapper)
	}
}

func (majSoul *MajSoul) readFastTestClientConn() {
	if majSoul.fastTestClientConn == nil {
		panic("fastTestClient Conn is nil")
	}
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

func (majSoul *MajSoul) OnGatewayReconnect(callback func(*message.ResLogin)) {
	majSoul.onGatewayReconnectCallBack = callback
}

func (majSoul *MajSoul) onGatewayReconnect() {
	if len(majSoul.AccessToken) == 0 {
		return
	}
	var err error
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err = majSoul.Oauth2Check(ctx, &message.ReqOauth2Check{AccessToken: majSoul.AccessToken})
		if err != nil {
			panic(fmt.Sprintf("gateway Oauth2Check error %v", err))
		}
	}
	var resLogin *message.ResLogin
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		resLogin, err = majSoul.Oauth2Login(ctx, &message.ReqOauth2Login{
			AccessToken: majSoul.AccessToken,
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
			Reconnect: false,
			RandomKey: majSoul.UUID,
			ClientVersion: &message.ClientVersionInfo{
				Resource: majSoul.Version.Version,
				Package:  "",
			},
			GenAccessToken:      false,
			CurrencyPlatforms:   []uint32{2, 6, 8, 10, 11},
			ClientVersionString: majSoul.Version.Web(),
		})
		if err != nil {
			panic(fmt.Sprintf("gateway Oauth2Login error %v", err))
		}
	}
	if majSoul.onGatewayReconnectCallBack != nil {
		majSoul.onGatewayReconnectCallBack(resLogin)
	}
}

func (majSoul *MajSoul) OnGameReconnect(callback func(*message.ResSyncGame)) {
	majSoul.onGameReconnectCallBack = callback
}

func (majSoul *MajSoul) onGameReconnect() {
	if len(majSoul.ConnectToken) == 0 {
		return
	}
	if len(majSoul.GameUuid) == 0 {
		return
	}

	var resSyncGame *message.ResSyncGame
	var err error
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		majSoul.GameInfo, err = majSoul.AuthGame(ctx, &message.ReqAuthGame{
			AccountId: majSoul.Account.AccountId,
			Token:     majSoul.ConnectToken,
			GameUuid:  majSoul.GameUuid,
		})
		if err != nil {
			logger.Error("majSoul AuthGame error.", zap.Error(err))
			return
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		resSyncGame, err = majSoul.SyncGame(ctx, &message.ReqSyncGame{RoundId: "-1"})
		if err != nil {
			logger.Error("majSoul SyncGame error.", zap.Error(err))
			return
		} else {
			logger.Debug("majSoul SyncGame.", zap.Reflect("resSyncGame", resSyncGame))
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if _, err = majSoul.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
			logger.Error("majSoul FetchGamePlayerState error.", zap.Error(err))
			return
		} else {
			logger.Debug("majSoul FetchGamePlayerState.")
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if _, err = majSoul.FinishSyncGame(ctx, &message.ReqCommon{}); err != nil {
			logger.Error("majSoul FinishSyncGame error.", zap.Error(err))
			return
		} else {
			logger.Debug("majSoul FinishSyncGame.")
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if _, err = majSoul.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
			logger.Error("majSoul FetchGamePlayerState error.", zap.Error(err))
			return
		} else {
			logger.Debug("majSoul FetchGamePlayerState.")
		}
	}

	if majSoul.onGameReconnectCallBack != nil {
		majSoul.onGameReconnectCallBack(resSyncGame)
	}
}

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
	if resLogin.Error == nil {
		majSoul.Account = resLogin.Account
		majSoul.AccessToken = resLogin.AccessToken
		if resLogin.GameInfo != nil {
			majSoul.ConnectToken = resLogin.GameInfo.ConnectToken
			majSoul.GameUuid = resLogin.GameInfo.GameUuid
		}
	}
	return resLogin, nil
}

func (majSoul *MajSoul) NotifyRoomGameStart(_ *MajSoul, notify *message.NotifyRoomGameStart) {
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := majSoul.ConnGame(ctx)
		if err != nil {
			panic(fmt.Sprintf("conn Game server failed error %v", err))
		}
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		var err error
		majSoul.GameInfo, err = majSoul.AuthGame(ctx, &message.ReqAuthGame{
			AccountId: majSoul.Account.AccountId,
			Token:     notify.ConnectToken,
			GameUuid:  notify.GameUuid,
		})
		if err != nil {
			logger.Error("majsoul NotifyRoomGameStart AuthGame error: ", zap.Error(err))
			return
		}
	}

	majSoul.ConnectToken = notify.ConnectToken
	majSoul.GameUuid = notify.GameUuid

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.EnterGame(ctx, &message.ReqCommon{})
		if err != nil {
			logger.Error("majsoul NotifyRoomGameStart EnterGame error:", zap.Error(err))
			return
		}
	}
}

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
