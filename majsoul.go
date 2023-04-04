// Package majsoul https://game.maj-soul.com/1/
package majsoul

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/constellation39/majsoul/logger"
	"go.uber.org/zap"

	"github.com/constellation39/majsoul/message"
	"google.golang.org/protobuf/proto"
)

const (
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54"
)

var (
	ErrorNoServerAvailable = errors.New("no server available")
)

type Implement interface {
	Notify // Notify 大厅通知下发
	Action // Action 游戏桌面内下发
}

type config struct {
	ServerProxy  string // 代理服务器地址(https)请求时，可以为空，为空时不使用代理
	GatewayProxy string // 代理网关服务器地址(wss)请求时，可以为空，为空时不使用代理
	GameProxy    string // 代理游戏服务器地址(wss)请求时，可以为空，为空时不使用代理
}

type ConfigOption func(*config)

func WithServerProxy(proxyAddress string) ConfigOption {
	return func(config *config) {
		config.ServerProxy = proxyAddress
	}
}

func WithGatewayProxy(proxyAddress string) ConfigOption {
	return func(config *config) {
		config.GatewayProxy = proxyAddress
	}
}

func WithGameProxy(proxyAddress string) ConfigOption {
	return func(config *config) {
		config.GameProxy = proxyAddress
	}
}

// Majsoul majsoul wsClient
type Majsoul struct {
	message.LobbyClient                         // message.LobbyClient 更多时候在大厅时调用的是该接口
	message.FastTestClient                      // message.FastTestClient 场景处于游戏桌面时调用该接口
	lobbyConn              *wsClient            // lobbyConn 是 message.LobbyClient 使用的连接
	fastTestConn           *wsClient            // fastTestConn 是 message.FastTestClient 使用的连接
	implement              Implement            // 使得程序可以以多态的方式调用 message.LobbyClient 或 message.FastTestClient 的接口
	UUID                   string               // uuid
	ServerAddress          *ServerAddress       // 连接到的服务器地址
	Request                *request             // 用于直接向http(s)请求
	Version                *Version             // 初始化时获取的版本信息
	Config                 *config              // Majsoul 初始化时使用的配置
	Account                *message.Account     // 该字段应在登录成功后访问
	GameInfo               *message.ResAuthGame // 该字段应在进入游戏桌面后访问
	accessToken            string               // 验证身份时使用 的 token
	connectToken           string               // 重连时使用的 token
	gameUuid               string               // 是否在游戏中

	cancelFunc                 context.CancelFunc
	onGatewayReconnectCallBack func(context.Context, *message.ResLogin)
	onGameReconnectCallBack    func(context.Context, *message.ResSyncGame)
}

// New Majsoul 是一个处理麻将游戏逻辑的结构体。要使用它，请先创建一个 Majsoul 对象，
func New(configOptions ...ConfigOption) *Majsoul {
	cfg := &config{}

	for _, configOption := range configOptions {
		configOption(cfg)
	}

	majsoul := &Majsoul{
		Config: cfg,
		UUID:   uuid(),
	}

	return majsoul
}

func (majsoul *Majsoul) setLobbyClient(client *wsClient) {
	if majsoul.lobbyConn != nil {
		majsoul.closeLobbyClient()
	}
	client.OnReconnect(majsoul.onGatewayReconnect)
	majsoul.lobbyConn = client
	majsoul.LobbyClient = message.NewLobbyClient(client)
}

func (majsoul *Majsoul) closeLobbyClient() {
	if majsoul.lobbyConn != nil {
		err := majsoul.lobbyConn.Close()
		if err != nil {
			logger.Error("majsoul closeCh lobby client error", zap.Error(err))
		}
		majsoul.lobbyConn = nil
	}
	if majsoul.lobbyConn != nil {
		majsoul.LobbyClient = nil
	}
}

func (majsoul *Majsoul) setFastTestClient(client *wsClient) {
	if majsoul.fastTestConn != nil {
		majsoul.closeFastTestClient()
	}
	client.OnReconnect(majsoul.onGameReconnect)
	majsoul.fastTestConn = client
	majsoul.FastTestClient = message.NewFastTestClient(client)
}

func (majsoul *Majsoul) closeFastTestClient() {
	if majsoul.fastTestConn != nil {
		err := majsoul.fastTestConn.Close()
		if err != nil {
			logger.Error("majsoul closeCh fast test client error", zap.Error(err))
		}
		majsoul.fastTestConn = nil
	}
	if majsoul.FastTestClient != nil {
		majsoul.FastTestClient = nil
	}
}

// Implement 每一个Majsoul对象都应该调用一次该方法，入参可以是Majsoul实例自己
func (majsoul *Majsoul) Implement(implement Implement) {
	majsoul.implement = implement
}

// TryConnect 尝试寻找可以使用的服务器
func (majsoul *Majsoul) TryConnect(ctx context.Context, ServerAddressList []*ServerAddress) (err error) {
	ctx, majsoul.cancelFunc = context.WithCancel(ctx)
	for _, serverAddress := range ServerAddressList {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var connUrl *url.URL
		connUrl, err = url.Parse(serverAddress.ServerAddress)

		if err != nil {
			continue
		}

		header := http.Header{}
		header.Add("Accept-Encoding", "gzip, deflate, br")
		header.Add("Accept-Language", "zh-CN,zh;q=0.9,ja;q=0.8,en;q=0.7,en-GB;q=0.6,en-US;q=0.5")
		header.Add("Cache-Control", "no-cache")
		header.Add("Host", connUrl.Host)
		header.Add("Origin", serverAddress.ServerAddress)
		header.Add("Pragma", "no-cache")
		header.Add("User-Agent", UserAgent)

		r := newRequest(serverAddress.ServerAddress, majsoul.Config.ServerProxy)
		_, err := r.Get(fmt.Sprintf("1/version.json?randv=%d", int(rand.Float32()*1000000000)+int(rand.Float32()*1000000000)))
		if err != nil {
			logger.Debug("majsoul request server failed.", zap.Reflect("serverAddress", serverAddress), zap.Reflect("config", majsoul.Config), zap.Error(err))
			continue
		}
		client := newWsClient(&wsConfig{
			ConnAddress:    serverAddress.GatewayAddress,
			ProxyAddress:   majsoul.Config.GatewayProxy,
			RequestHeaders: header,
		})
		err = client.Connect(ctx)
		if err != nil {
			logger.Debug("majsoul connect server failed.", zap.Reflect("serverAddress", serverAddress), zap.Reflect("config", majsoul.Config), zap.Error(err))
			continue
		}
		majsoul.ServerAddress = serverAddress
		majsoul.Request = r
		majsoul.setLobbyClient(client)

		err = majsoul.initVersion(ctx)
		if err != nil {
			logger.Debug("majsoul init version failed.", zap.Reflect("serverAddress", serverAddress), zap.Reflect("config", majsoul.Config), zap.Error(err))
			continue
		}

		go majsoul.heatbeat(ctx)
		go majsoul.receiveConn(ctx)

		return nil
	}
	return ErrorNoServerAvailable
}

func (majsoul *Majsoul) Close() {
	majsoul.cancelFunc()
}

// initVersion 获取版本
func (majsoul *Majsoul) initVersion(ctx context.Context) (err error) {
	majsoul.Version, err = majsoul.version()
	return
}

// ConnGame 连接到对局服务器
func (majsoul *Majsoul) ConnGame(ctx context.Context) (err error) {
	connUrl, err := url.Parse(majsoul.ServerAddress.GameAddress)
	if err != nil {
		logger.Error("majsoul failed to parse GameAddress: ", zap.String("GameAddress", majsoul.ServerAddress.GameAddress), zap.Error(err))
	}

	header := http.Header{}
	header.Add("Accept-Encoding", "gzip, deflate, br")
	header.Add("Accept-Language", "zh-CN,zh;q=0.9,ja;q=0.8,en;q=0.7,en-GB;q=0.6,en-US;q=0.5")
	header.Add("Cache-Control", "no-cache")
	header.Add("Host", connUrl.Host)
	header.Add("Origin", majsoul.ServerAddress.ServerAddress)
	header.Add("Pragma", "no-cache")
	header.Add("User-Agent", UserAgent)

	clinet := newWsClient(&wsConfig{
		ConnAddress:    majsoul.ServerAddress.GameAddress,
		ProxyAddress:   majsoul.Config.GameProxy,
		RequestHeaders: header,
	})
	err = clinet.Connect(ctx)
	if err != nil {
		logger.Error("majsoul failed to connect to GameServer: ", zap.String("GameAddress", majsoul.ServerAddress.GameAddress), zap.Error(err))
		return
	}

	majsoul.setFastTestClient(clinet)
	go majsoul.receiveGame(ctx)
	return
}

// ReConnGame reconnects to the game server.
func (majsoul *Majsoul) ReConnGame(ctx context.Context, resLogin *message.ResLogin) error {
	if resLogin.Account == nil || resLogin.Account.RoomId == 0 {
		return nil
	}

	if err := majsoul.ConnGame(ctx); err != nil {
		return err
	}

	var err error
	majsoul.GameInfo, err = majsoul.AuthGame(ctx, &message.ReqAuthGame{
		AccountId: majsoul.Account.AccountId,
		Token:     resLogin.GameInfo.ConnectToken,
		GameUuid:  resLogin.GameInfo.GameUuid,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate game connection: %v", err)
	}

	if resSyncGame, err := majsoul.SyncGame(ctx, &message.ReqSyncGame{RoundId: "-1"}); err != nil {
		return fmt.Errorf("failed to sync game state: %v", err)
	} else {
		logger.Debug("majsoul SyncGame.", zap.Reflect("resSyncGame", resSyncGame))
	}

	if _, err := majsoul.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
		return fmt.Errorf("failed to fetch game player state: %v", err)
	} else {
		logger.Debug("majsoul FetchGamePlayerState.")
	}

	if _, err := majsoul.FinishSyncGame(ctx, &message.ReqCommon{}); err != nil {
		return fmt.Errorf("failed to fetch game player state: %v", err)
	} else {
		logger.Debug("majsoul FinishSyncGame.")
	}

	if _, err := majsoul.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
		return fmt.Errorf("failed to fetch game player state after syncing: %v", err)
	} else {
		logger.Debug("majsoul FetchGamePlayerState.")
	}

	return nil
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

func (majsoul *Majsoul) version() (*Version, error) {
	// var version_url = "version.json?randv="+Math.floor(Math.random() * 1000000000).toString()+Math.floor(Math.random() * 1000000000).toString()
	r := int(rand.Float32()*1e9) + int(rand.Float32()*1e9)
	body, err := majsoul.Request.Get(fmt.Sprintf("1/version.json?randv=%d", r))
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

func (majsoul *Majsoul) heatbeat(ctx context.Context) {
	// Gateway 心跳包 5 秒一次
	t5 := time.NewTicker(time.Second * 5)
	// Game 心跳包 2 秒一次
	t2 := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t5.C:
			if majsoul.fastTestConn != nil {
				continue
			}
			_, err := majsoul.Heatbeat(ctx, &message.ReqHeatBeat{})
			if err != nil {
				logger.Error("majsoul heatbeat error:", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
		case <-t2.C:
			if majsoul.fastTestConn == nil {
				continue
			}
			_, err := majsoul.CheckNetworkDelay(ctx, &message.ReqCommon{})
			if err != nil {
				logger.Error("majsoul checkNetworkDelay error:", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
		}
	}
}

func (majsoul *Majsoul) receiveConn(ctx context.Context) {
	if majsoul.lobbyConn == nil {
		logger.Panic("majsoul lobbyConn is nil")
	}
	receive := majsoul.lobbyConn.Receive()
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-receive:
			if !ok {
				logger.Debug("majsoul lobbyConn closeCh")
				return
			}
			majsoul.handleNotify(ctx, data)
		}
	}
}

func (majsoul *Majsoul) receiveGame(ctx context.Context) {
	if majsoul.fastTestConn == nil {
		logger.Panic("majsoul fastTestConn is nil")
	}
	receive := majsoul.fastTestConn.Receive()
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-receive:
			if !ok {
				logger.Debug("majsoul fastTestConn closeCh")
				return
			}
			majsoul.handleNotify(ctx, data)
		}
	}
}

func (majsoul *Majsoul) handleNotify(ctx context.Context, data proto.Message) {
	if majsoul.implement == nil {
		logger.Panic("majsoul implement is null, please set majsoul.implement first by majsoul.Implement func")
	}

	switch notify := data.(type) {
	case *message.NotifyCaptcha:
		logger.Debug("majsoul NotifyCaptcha.", zap.Reflect("data", notify))
		majsoul.NotifyCaptcha(ctx, notify)
		majsoul.implement.NotifyCaptcha(ctx, notify)
	case *message.NotifyRoomGameStart:
		logger.Debug("majsoul NotifyRoomGameStart.", zap.Reflect("data", notify))
		majsoul.NotifyRoomGameStart(ctx, notify)
		majsoul.implement.NotifyRoomGameStart(ctx, notify)
	case *message.NotifyMatchGameStart:
		logger.Debug("majsoul NotifyMatchGameStart.", zap.Reflect("data", notify))
		majsoul.NotifyMatchGameStart(ctx, notify)
		majsoul.implement.NotifyMatchGameStart(ctx, notify)
	case *message.NotifyRoomPlayerReady:
		logger.Debug("majsoul NotifyRoomPlayerReady.", zap.Reflect("data", notify))
		majsoul.NotifyRoomPlayerReady(ctx, notify)
		majsoul.implement.NotifyRoomPlayerReady(ctx, notify)
	case *message.NotifyRoomPlayerDressing:
		logger.Debug("majsoul NotifyRoomPlayerDressing.", zap.Reflect("data", notify))
		majsoul.NotifyRoomPlayerDressing(ctx, notify)
		majsoul.implement.NotifyRoomPlayerDressing(ctx, notify)
	case *message.NotifyRoomPlayerUpdate:
		logger.Debug("majsoul NotifyRoomPlayerUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyRoomPlayerUpdate(ctx, notify)
		majsoul.implement.NotifyRoomPlayerUpdate(ctx, notify)
	case *message.NotifyRoomKickOut:
		logger.Debug("majsoul NotifyRoomKickOut.", zap.Reflect("data", notify))
		majsoul.NotifyRoomKickOut(ctx, notify)
		majsoul.implement.NotifyRoomKickOut(ctx, notify)
	case *message.NotifyFriendStateChange:
		logger.Debug("majsoul NotifyFriendStateChange.", zap.Reflect("data", notify))
		majsoul.NotifyFriendStateChange(ctx, notify)
		majsoul.implement.NotifyFriendStateChange(ctx, notify)
	case *message.NotifyFriendViewChange:
		logger.Debug("majsoul NotifyFriendViewChange.", zap.Reflect("data", notify))
		majsoul.NotifyFriendViewChange(ctx, notify)
		majsoul.implement.NotifyFriendViewChange(ctx, notify)
	case *message.NotifyFriendChange:
		logger.Debug("majsoul NotifyFriendChange.", zap.Reflect("data", notify))
		majsoul.NotifyFriendChange(ctx, notify)
		majsoul.implement.NotifyFriendChange(ctx, notify)
	case *message.NotifyNewFriendApply:
		logger.Debug("majsoul NotifyNewFriendApply.", zap.Reflect("data", notify))
		majsoul.NotifyNewFriendApply(ctx, notify)
		majsoul.implement.NotifyNewFriendApply(ctx, notify)
	case *message.NotifyClientMessage:
		logger.Debug("majsoul NotifyClientMessage.", zap.Reflect("data", notify))
		majsoul.NotifyClientMessage(ctx, notify)
		majsoul.implement.NotifyClientMessage(ctx, notify)
	case *message.NotifyAccountUpdate:
		logger.Debug("majsoul NotifyAccountUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyAccountUpdate(ctx, notify)
		majsoul.implement.NotifyAccountUpdate(ctx, notify)
	case *message.NotifyAnotherLogin:
		logger.Debug("majsoul NotifyAnotherLogin.", zap.Reflect("data", notify))
		majsoul.NotifyAnotherLogin(ctx, notify)
		majsoul.implement.NotifyAnotherLogin(ctx, notify)
	case *message.NotifyAccountLogout:
		logger.Debug("majsoul NotifyAccountLogout.", zap.Reflect("data", notify))
		majsoul.NotifyAccountLogout(ctx, notify)
		majsoul.implement.NotifyAccountLogout(ctx, notify)
	case *message.NotifyAnnouncementUpdate:
		logger.Debug("majsoul NotifyAnnouncementUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyAnnouncementUpdate(ctx, notify)
		majsoul.implement.NotifyAnnouncementUpdate(ctx, notify)
	case *message.NotifyNewMail:
		logger.Debug("majsoul NotifyNewMail.", zap.Reflect("data", notify))
		majsoul.NotifyNewMail(ctx, notify)
		majsoul.implement.NotifyNewMail(ctx, notify)
	case *message.NotifyDeleteMail:
		logger.Debug("majsoul NotifyDeleteMail.", zap.Reflect("data", notify))
		majsoul.NotifyDeleteMail(ctx, notify)
		majsoul.implement.NotifyDeleteMail(ctx, notify)
	case *message.NotifyReviveCoinUpdate:
		logger.Debug("majsoul NotifyReviveCoinUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyReviveCoinUpdate(ctx, notify)
		majsoul.implement.NotifyReviveCoinUpdate(ctx, notify)
	case *message.NotifyDailyTaskUpdate:
		logger.Debug("majsoul NotifyDailyTaskUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyDailyTaskUpdate(ctx, notify)
		majsoul.implement.NotifyDailyTaskUpdate(ctx, notify)
	case *message.NotifyActivityTaskUpdate:
		logger.Debug("majsoul NotifyActivityTaskUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyActivityTaskUpdate(ctx, notify)
		majsoul.implement.NotifyActivityTaskUpdate(ctx, notify)
	case *message.NotifyActivityPeriodTaskUpdate:
		logger.Debug("majsoul NotifyActivityPeriodTaskUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyActivityPeriodTaskUpdate(ctx, notify)
		majsoul.implement.NotifyActivityPeriodTaskUpdate(ctx, notify)
	case *message.NotifyAccountRandomTaskUpdate:
		logger.Debug("majsoul NotifyAccountRandomTaskUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyAccountRandomTaskUpdate(ctx, notify)
		majsoul.implement.NotifyAccountRandomTaskUpdate(ctx, notify)
	case *message.NotifyActivitySegmentTaskUpdate:
		logger.Debug("majsoul NotifyActivitySegmentTaskUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyActivitySegmentTaskUpdate(ctx, notify)
		majsoul.implement.NotifyActivitySegmentTaskUpdate(ctx, notify)
	case *message.NotifyActivityUpdate:
		logger.Debug("majsoul NotifyActivityUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyActivityUpdate(ctx, notify)
		majsoul.implement.NotifyActivityUpdate(ctx, notify)
	case *message.NotifyAccountChallengeTaskUpdate:
		logger.Debug("majsoul NotifyAccountChallengeTaskUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyAccountChallengeTaskUpdate(ctx, notify)
		majsoul.implement.NotifyAccountChallengeTaskUpdate(ctx, notify)
	case *message.NotifyNewComment:
		logger.Debug("majsoul NotifyNewComment.", zap.Reflect("data", notify))
		majsoul.NotifyNewComment(ctx, notify)
		majsoul.implement.NotifyNewComment(ctx, notify)
	case *message.NotifyRollingNotice:
		logger.Debug("majsoul NotifyRollingNotice.", zap.Reflect("data", notify))
		majsoul.NotifyRollingNotice(ctx, notify)
		majsoul.implement.NotifyRollingNotice(ctx, notify)
	case *message.NotifyGiftSendRefresh:
		logger.Debug("majsoul NotifyGiftSendRefresh.", zap.Reflect("data", notify))
		majsoul.NotifyGiftSendRefresh(ctx, notify)
		majsoul.implement.NotifyGiftSendRefresh(ctx, notify)
	case *message.NotifyShopUpdate:
		logger.Debug("majsoul NotifyShopUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyShopUpdate(ctx, notify)
		majsoul.implement.NotifyShopUpdate(ctx, notify)
	case *message.NotifyVipLevelChange:
		logger.Debug("majsoul NotifyVipLevelChange.", zap.Reflect("data", notify))
		majsoul.NotifyVipLevelChange(ctx, notify)
		majsoul.implement.NotifyVipLevelChange(ctx, notify)
	case *message.NotifyServerSetting:
		logger.Debug("majsoul NotifyServerSetting.", zap.Reflect("data", notify))
		majsoul.NotifyServerSetting(ctx, notify)
		majsoul.implement.NotifyServerSetting(ctx, notify)
	case *message.NotifyPayResult:
		logger.Debug("majsoul NotifyPayResult.", zap.Reflect("data", notify))
		majsoul.NotifyPayResult(ctx, notify)
		majsoul.implement.NotifyPayResult(ctx, notify)
	case *message.NotifyCustomContestAccountMsg:
		logger.Debug("majsoul NotifyCustomContestAccountMsg.", zap.Reflect("data", notify))
		majsoul.NotifyCustomContestAccountMsg(ctx, notify)
		majsoul.implement.NotifyCustomContestAccountMsg(ctx, notify)
	case *message.NotifyCustomContestSystemMsg:
		logger.Debug("majsoul NotifyCustomContestSystemMsg.", zap.Reflect("data", notify))
		majsoul.NotifyCustomContestSystemMsg(ctx, notify)
		majsoul.implement.NotifyCustomContestSystemMsg(ctx, notify)
	case *message.NotifyMatchTimeout:
		logger.Debug("majsoul NotifyMatchTimeout.", zap.Reflect("data", notify))
		majsoul.NotifyMatchTimeout(ctx, notify)
		majsoul.implement.NotifyMatchTimeout(ctx, notify)
	case *message.NotifyCustomContestState:
		logger.Debug("majsoul NotifyCustomContestState.", zap.Reflect("data", notify))
		majsoul.NotifyCustomContestState(ctx, notify)
		majsoul.implement.NotifyCustomContestState(ctx, notify)
	case *message.NotifyActivityChange:
		logger.Debug("majsoul NotifyActivityChange.", zap.Reflect("data", notify))
		majsoul.NotifyActivityChange(ctx, notify)
		majsoul.implement.NotifyActivityChange(ctx, notify)
	case *message.NotifyAFKResult:
		logger.Debug("majsoul NotifyAFKResult.", zap.Reflect("data", notify))
		majsoul.NotifyAFKResult(ctx, notify)
		majsoul.implement.NotifyAFKResult(ctx, notify)
	case *message.NotifyGameFinishRewardV2:
		logger.Debug("majsoul NotifyGameFinishRewardV2.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishRewardV2(ctx, notify)
		majsoul.implement.NotifyGameFinishRewardV2(ctx, notify)
	case *message.NotifyActivityRewardV2:
		logger.Debug("majsoul NotifyActivityRewardV2.", zap.Reflect("data", notify))
		majsoul.NotifyActivityRewardV2(ctx, notify)
		majsoul.implement.NotifyActivityRewardV2(ctx, notify)
	case *message.NotifyActivityPointV2:
		logger.Debug("majsoul NotifyActivityPointV2.", zap.Reflect("data", notify))
		majsoul.NotifyActivityPointV2(ctx, notify)
		majsoul.implement.NotifyActivityPointV2(ctx, notify)
	case *message.NotifyLeaderboardPointV2:
		logger.Debug("majsoul NotifyLeaderboardPointV2.", zap.Reflect("data", notify))
		majsoul.NotifyLeaderboardPointV2(ctx, notify)
		majsoul.implement.NotifyLeaderboardPointV2(ctx, notify)
	case *message.NotifyNewGame:
		logger.Debug("majsoul NotifyNewGame.", zap.Reflect("data", notify))
		majsoul.NotifyNewGame(ctx, notify)
		majsoul.implement.NotifyNewGame(ctx, notify)
	case *message.NotifyPlayerLoadGameReady:
		logger.Debug("majsoul NotifyPlayerLoadGameReady.", zap.Reflect("data", notify))
		majsoul.NotifyPlayerLoadGameReady(ctx, notify)
		majsoul.implement.NotifyPlayerLoadGameReady(ctx, notify)
	case *message.NotifyGameBroadcast:
		logger.Debug("majsoul NotifyGameBroadcast.", zap.Reflect("data", notify))
		majsoul.NotifyGameBroadcast(ctx, notify)
		majsoul.implement.NotifyGameBroadcast(ctx, notify)
	case *message.NotifyGameEndResult:
		logger.Debug("majsoul NotifyGameEndResult.", zap.Reflect("data", notify))
		majsoul.NotifyGameEndResult(ctx, notify)
		majsoul.implement.NotifyGameEndResult(ctx, notify)
	case *message.NotifyGameTerminate:
		logger.Debug("majsoul NotifyGameTerminate.", zap.Reflect("data", notify))
		majsoul.NotifyGameTerminate(ctx, notify)
		majsoul.implement.NotifyGameTerminate(ctx, notify)
	case *message.NotifyPlayerConnectionState:
		logger.Debug("majsoul NotifyPlayerConnectionState.", zap.Reflect("data", notify))
		majsoul.NotifyPlayerConnectionState(ctx, notify)
		majsoul.implement.NotifyPlayerConnectionState(ctx, notify)
	case *message.NotifyAccountLevelChange:
		logger.Debug("majsoul NotifyAccountLevelChange.", zap.Reflect("data", notify))
		majsoul.NotifyAccountLevelChange(ctx, notify)
		majsoul.implement.NotifyAccountLevelChange(ctx, notify)
	case *message.NotifyGameFinishReward:
		logger.Debug("majsoul NotifyGameFinishReward.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishReward(ctx, notify)
		majsoul.implement.NotifyGameFinishReward(ctx, notify)
	case *message.NotifyActivityReward:
		logger.Debug("majsoul NotifyActivityReward.", zap.Reflect("data", notify))
		majsoul.NotifyActivityReward(ctx, notify)
		majsoul.implement.NotifyActivityReward(ctx, notify)
	case *message.NotifyActivityPoint:
		logger.Debug("majsoul NotifyActivityPoint.", zap.Reflect("data", notify))
		majsoul.NotifyActivityPoint(ctx, notify)
		majsoul.implement.NotifyActivityPoint(ctx, notify)
	case *message.NotifyLeaderboardPoint:
		logger.Debug("majsoul NotifyLeaderboardPoint.", zap.Reflect("data", notify))
		majsoul.NotifyLeaderboardPoint(ctx, notify)
		majsoul.implement.NotifyLeaderboardPoint(ctx, notify)
	case *message.NotifyGamePause:
		logger.Debug("majsoul NotifyGamePause.", zap.Reflect("data", notify))
		majsoul.NotifyGamePause(ctx, notify)
		majsoul.implement.NotifyGamePause(ctx, notify)
	case *message.NotifyEndGameVote:
		logger.Debug("majsoul NotifyEndGameVote.", zap.Reflect("data", notify))
		majsoul.NotifyEndGameVote(ctx, notify)
		majsoul.implement.NotifyEndGameVote(ctx, notify)
	case *message.NotifyObserveData:
		logger.Debug("majsoul NotifyObserveData.", zap.Reflect("data", notify))
		majsoul.NotifyObserveData(ctx, notify)
		majsoul.implement.NotifyObserveData(ctx, notify)
	case *message.NotifyRoomPlayerReady_AccountReadyState:
		logger.Debug("majsoul NotifyRoomPlayerReady_AccountReadyState.", zap.Reflect("data", notify))
		majsoul.NotifyRoomPlayerReady_AccountReadyState(ctx, notify)
		majsoul.implement.NotifyRoomPlayerReady_AccountReadyState(ctx, notify)
	case *message.NotifyRoomPlayerDressing_AccountDressingState:
		logger.Debug("majsoul NotifyRoomPlayerDressing_AccountDressingState.", zap.Reflect("data", notify))
		majsoul.NotifyRoomPlayerDressing_AccountDressingState(ctx, notify)
		majsoul.implement.NotifyRoomPlayerDressing_AccountDressingState(ctx, notify)
	case *message.NotifyAnnouncementUpdate_AnnouncementUpdate:
		logger.Debug("majsoul NotifyAnnouncementUpdate_AnnouncementUpdate.", zap.Reflect("data", notify))
		majsoul.NotifyAnnouncementUpdate_AnnouncementUpdate(ctx, notify)
		majsoul.implement.NotifyAnnouncementUpdate_AnnouncementUpdate(ctx, notify)
	case *message.NotifyActivityUpdate_FeedActivityData:
		logger.Debug("majsoul NotifyActivityUpdate_FeedActivityData.", zap.Reflect("data", notify))
		majsoul.NotifyActivityUpdate_FeedActivityData(ctx, notify)
		majsoul.implement.NotifyActivityUpdate_FeedActivityData(ctx, notify)
	case *message.NotifyActivityUpdate_FeedActivityData_CountWithTimeData:
		logger.Debug("majsoul NotifyActivityUpdate_FeedActivityData_CountWithTimeData.", zap.Reflect("data", notify))
		majsoul.NotifyActivityUpdate_FeedActivityData_CountWithTimeData(ctx, notify)
		majsoul.implement.NotifyActivityUpdate_FeedActivityData_CountWithTimeData(ctx, notify)
	case *message.NotifyActivityUpdate_FeedActivityData_GiftBoxData:
		logger.Debug("majsoul NotifyActivityUpdate_FeedActivityData_GiftBoxData.", zap.Reflect("data", notify))
		majsoul.NotifyActivityUpdate_FeedActivityData_GiftBoxData(ctx, notify)
		majsoul.implement.NotifyActivityUpdate_FeedActivityData_GiftBoxData(ctx, notify)
	case *message.NotifyPayResult_ResourceModify:
		logger.Debug("majsoul NotifyPayResult_ResourceModify.", zap.Reflect("data", notify))
		majsoul.NotifyPayResult_ResourceModify(ctx, notify)
		majsoul.implement.NotifyPayResult_ResourceModify(ctx, notify)
	case *message.NotifyGameFinishRewardV2_LevelChange:
		logger.Debug("majsoul NotifyGameFinishRewardV2_LevelChange.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishRewardV2_LevelChange(ctx, notify)
		majsoul.implement.NotifyGameFinishRewardV2_LevelChange(ctx, notify)
	case *message.NotifyGameFinishRewardV2_MatchChest:
		logger.Debug("majsoul NotifyGameFinishRewardV2_MatchChest.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishRewardV2_MatchChest(ctx, notify)
		majsoul.implement.NotifyGameFinishRewardV2_MatchChest(ctx, notify)
	case *message.NotifyGameFinishRewardV2_MainCharacter:
		logger.Debug("majsoul NotifyGameFinishRewardV2_MainCharacter.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishRewardV2_MainCharacter(ctx, notify)
		majsoul.implement.NotifyGameFinishRewardV2_MainCharacter(ctx, notify)
	case *message.NotifyGameFinishRewardV2_CharacterGift:
		logger.Debug("majsoul NotifyGameFinishRewardV2_CharacterGift.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishRewardV2_CharacterGift(ctx, notify)
		majsoul.implement.NotifyGameFinishRewardV2_CharacterGift(ctx, notify)
	case *message.NotifyActivityRewardV2_ActivityReward:
		logger.Debug("majsoul NotifyActivityRewardV2_ActivityReward.", zap.Reflect("data", notify))
		majsoul.NotifyActivityRewardV2_ActivityReward(ctx, notify)
		majsoul.implement.NotifyActivityRewardV2_ActivityReward(ctx, notify)
	case *message.NotifyActivityPointV2_ActivityPoint:
		logger.Debug("majsoul NotifyActivityPointV2_ActivityPoint.", zap.Reflect("data", notify))
		majsoul.NotifyActivityPointV2_ActivityPoint(ctx, notify)
		majsoul.implement.NotifyActivityPointV2_ActivityPoint(ctx, notify)
	case *message.NotifyLeaderboardPointV2_LeaderboardPoint:
		logger.Debug("majsoul NotifyLeaderboardPointV2_LeaderboardPoint.", zap.Reflect("data", notify))
		majsoul.NotifyLeaderboardPointV2_LeaderboardPoint(ctx, notify)
		majsoul.implement.NotifyLeaderboardPointV2_LeaderboardPoint(ctx, notify)
	case *message.NotifyGameFinishReward_LevelChange:
		logger.Debug("majsoul NotifyGameFinishReward_LevelChange.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishReward_LevelChange(ctx, notify)
		majsoul.implement.NotifyGameFinishReward_LevelChange(ctx, notify)
	case *message.NotifyGameFinishReward_MatchChest:
		logger.Debug("majsoul NotifyGameFinishReward_MatchChest.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishReward_MatchChest(ctx, notify)
		majsoul.implement.NotifyGameFinishReward_MatchChest(ctx, notify)
	case *message.NotifyGameFinishReward_MainCharacter:
		logger.Debug("majsoul NotifyGameFinishReward_MainCharacter.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishReward_MainCharacter(ctx, notify)
		majsoul.implement.NotifyGameFinishReward_MainCharacter(ctx, notify)
	case *message.NotifyGameFinishReward_CharacterGift:
		logger.Debug("majsoul NotifyGameFinishReward_CharacterGift.", zap.Reflect("data", notify))
		majsoul.NotifyGameFinishReward_CharacterGift(ctx, notify)
		majsoul.implement.NotifyGameFinishReward_CharacterGift(ctx, notify)
	case *message.NotifyActivityReward_ActivityReward:
		logger.Debug("majsoul NotifyActivityReward_ActivityReward.", zap.Reflect("data", notify))
		majsoul.NotifyActivityReward_ActivityReward(ctx, notify)
		majsoul.implement.NotifyActivityReward_ActivityReward(ctx, notify)
	case *message.NotifyActivityPoint_ActivityPoint:
		logger.Debug("majsoul NotifyActivityPoint_ActivityPoint.", zap.Reflect("data", notify))
		majsoul.NotifyActivityPoint_ActivityPoint(ctx, notify)
		majsoul.implement.NotifyActivityPoint_ActivityPoint(ctx, notify)
	case *message.NotifyLeaderboardPoint_LeaderboardPoint:
		logger.Debug("majsoul NotifyLeaderboardPoint_LeaderboardPoint.", zap.Reflect("data", notify))
		majsoul.NotifyLeaderboardPoint_LeaderboardPoint(ctx, notify)
		majsoul.implement.NotifyLeaderboardPoint_LeaderboardPoint(ctx, notify)
	case *message.NotifyEndGameVote_VoteResult:
		logger.Debug("majsoul NotifyEndGameVote_VoteResult.", zap.Reflect("data", notify))
		majsoul.NotifyEndGameVote_VoteResult(ctx, notify)
		majsoul.implement.NotifyEndGameVote_VoteResult(ctx, notify)
	case *message.PlayerLeaving:
		logger.Debug("majsoul PlayerLeaving.", zap.Reflect("data", notify))
		majsoul.PlayerLeaving(ctx, notify)
		majsoul.implement.PlayerLeaving(ctx, notify)
	case *message.ActionPrototype:
		logger.Debug("majsoul ActionPrototype.", zap.Reflect("data", notify))
		// majsoul.ActionPrototype(ctx, notify)
		majsoul.implement.ActionPrototype(ctx, notify)
	default:
		logger.Info("majsoul unknown notify type", zap.Reflect("notify", notify))
	}
}

func uuid() string {
	const charSet = "0123456789abcdefghijklmnopqrstuvwxyz"
	csl := len(charSet)
	b := make([]byte, 36)
	for i := 0; i < 36; i++ {
		if i == 7 || i == 12 || i == 17 || i == 22 {
			b[i] = '-'
			continue
		}
		b[i] = charSet[rand.Intn(csl)]
	}
	return string(b)
}

// hashPassword password with hmac sha256
// return hash string
func hashPassword(data string) string {
	hash := hmac.New(sha256.New, []byte("lailai"))
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

func ErrorString(err *message.Error) (msg string) {
	switch err.Code {
	case 0:
		msg = ""
	case 103:
		msg = "维护中，服务器暂未开放"
	case 109:
		msg = "授权出错，登入已过期，请重新登入"
	case 1002:
		msg = "账号不存在，请先注册"
	case 1003:
		msg = "账号或密码错误"
	default:
		msg = fmt.Sprintf("unknown code (%d), message:%s",
			err.Code, strings.Join(err.StrParams, " "))
	}
	return
}

func (majsoul *Majsoul) OnGatewayReconnect(callback func(context.Context, *message.ResLogin)) {
	majsoul.onGatewayReconnectCallBack = callback
}

// onGatewayReconnect 断线重连
// 这个callbreak内应该先与服务器进行验权，在进行接下来的交互
// 拥有一个默认实现
func (majsoul *Majsoul) onGatewayReconnect(ctx context.Context) {
	if len(majsoul.accessToken) == 0 {
		logger.Debug("accessToken is nil")
		return
	}
	resOauth2Check, err := majsoul.Oauth2Check(ctx, &message.ReqOauth2Check{AccessToken: majsoul.accessToken})
	if err != nil {
		logger.Error("majsoul Oauth2Check error.", zap.Error(err))
	}
	logger.Debug("majsoul Oauth2Check.", zap.Reflect("resOauth2Check", resOauth2Check))
	resLogin, err := majsoul.Oauth2Login(ctx, &message.ReqOauth2Login{
		AccessToken: majsoul.accessToken,
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
		RandomKey: majsoul.UUID,
		ClientVersion: &message.ClientVersionInfo{
			Resource: majsoul.Version.Version,
			Package:  "",
		},
		GenAccessToken:      false,
		CurrencyPlatforms:   []uint32{2, 6, 8, 10, 11},
		ClientVersionString: majsoul.Version.Web(),
	})
	if err != nil {
		logger.Error("majsoul Oauth2Login error.", zap.Error(err))
	}
	logger.Debug("majsoul Oauth2Login.", zap.Reflect("resLogin", resLogin))
	if majsoul.onGatewayReconnectCallBack != nil {
		majsoul.onGatewayReconnectCallBack(ctx, resLogin)
	}
}

func (majsoul *Majsoul) OnGameReconnect(callback func(context.Context, *message.ResSyncGame)) {
	majsoul.onGameReconnectCallBack = callback
}

func (majsoul *Majsoul) onGameReconnect(ctx context.Context) {
	if len(majsoul.connectToken) == 0 {
		logger.Debug("connectToken is nil")
		return
	}
	if len(majsoul.gameUuid) == 0 {
		logger.Debug("gameUuid is nil")
		return
	}
	var err error
	majsoul.GameInfo, err = majsoul.AuthGame(ctx, &message.ReqAuthGame{
		AccountId: majsoul.Account.AccountId,
		Token:     majsoul.connectToken,
		GameUuid:  majsoul.gameUuid,
	})
	if err != nil {
		logger.Error("majsoul AuthGame error.", zap.Error(err))
		return
	}

	resSyncGame, err := majsoul.SyncGame(ctx, &message.ReqSyncGame{RoundId: "-1"})
	if err != nil {
		logger.Error("majsoul SyncGame error.", zap.Error(err))
		return
	} else {
		logger.Debug("majsoul SyncGame.", zap.Reflect("resSyncGame", resSyncGame))
	}

	if _, err := majsoul.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
		logger.Error("majsoul FetchGamePlayerState error.", zap.Error(err))
		return
	} else {
		logger.Debug("majsoul FetchGamePlayerState.")
	}

	if _, err := majsoul.FinishSyncGame(ctx, &message.ReqCommon{}); err != nil {
		logger.Error("majsoul FinishSyncGame error.", zap.Error(err))
		return
	} else {
		logger.Debug("majsoul FinishSyncGame.")
	}

	if _, err := majsoul.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
		logger.Error("majsoul FetchGamePlayerState error.", zap.Error(err))
		return
	} else {
		logger.Debug("majsoul FetchGamePlayerState.")
	}

	if majsoul.onGameReconnectCallBack != nil {
		// 这里没有加密
		// if resSyncGame.GameRestore != nil {
		// 	if resSyncGame.GameRestore.Actions != nil {
		// 		for _, action := range resSyncGame.GameRestore.Actions {
		// 			DecodeActionPrototype(action)
		// 		}
		// 	}
		// }
		majsoul.onGameReconnectCallBack(ctx, resSyncGame)
	}
}

// Login 登录，这是一个额外实现，并不属于 proto 或者 GRPC 的定义中
func (majsoul *Majsoul) Login(ctx context.Context, account, password string) (*message.ResLogin, error) {
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
	resLogin, err := majsoul.LobbyClient.Login(ctx, &message.ReqLogin{
		Account:   account,
		Password:  hashPassword(password),
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
		RandomKey: majsoul.UUID,
		ClientVersion: &message.ClientVersionInfo{
			Resource: majsoul.Version.Version,
			Package:  "",
		},
		GenAccessToken:    true,
		CurrencyPlatforms: []uint32{2, 6, 8, 10, 11},
		// 电话1 邮箱0
		Type:                t,
		Version:             0,
		ClientVersionString: majsoul.Version.Web(),
	})
	if err != nil {
		return nil, err
	}
	if resLogin.Error == nil {
		majsoul.Account = resLogin.Account
		majsoul.accessToken = resLogin.AccessToken
		if resLogin.GameInfo != nil {
			majsoul.connectToken = resLogin.GameInfo.ConnectToken
			majsoul.gameUuid = resLogin.GameInfo.GameUuid
		}
	}
	return resLogin, nil
}
