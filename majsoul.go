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
	MsgTypeNotify   uint8 = 1 // 通知
	MsgTypeRequest  uint8 = 2 // 请求
	MsgTypeResponse uint8 = 3 // 回复

	ActionDiscard = 1  // 出牌
	ActionChi     = 2  // 吃
	ActionPon     = 3  // 碰
	ActionAnKAN   = 4  // 暗槓
	ActionMinKan  = 5  // 明槓
	ActionKaKan   = 6  // 加槓
	ActionRiichi  = 7  // 立直
	ActionTsumo   = 8  // 自摸
	ActionRon     = 9  // 栄和
	ActionKuku    = 10 // 九九流局
	ActionKita    = 11 // 北
	ActionPass    = 12 // 見逃

	NotifyChi   = 0 // 吃
	NotifyPon   = 1 // 碰
	NotifyKan   = 2 // 杠
	NotifyAnKan = 3 // 暗杠
	NotifyKaKan = 4 // 加杠

	EBakaze = 0 // 东风
	SBakaze = 1 // 南风
	WBakaze = 2 // 西风
	NBakaze = 3 // 北风

	Toncha = 0 // 東家
	Nancha = 1 // 南家
	ShaCha = 2 // 西家
	Peicha = 3 // 北家

	Kyoku1 = 0 // 第1局
	Kyoku2 = 1 // 第2局
	Kyoku3 = 2 // 第3局
	Kyoku4 = 3 // 第4局
)

const (
	charSet   = "0123456789abcdefghijklmnopqrstuvwxyz"
	uuidFile  = ".UUID"
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54"
)

var (
	ErrorNoServerAvailable = errors.New("no server available")
	ErrorShutdownSignal    = errors.New("receive shutdown signal")
)

type Implement interface {
	IFNotify // IFNotify 大厅通知下发
	IFAction // IFAction  游戏桌面内下发
}

type Config struct {
	ServerAddressList []*ServerAddress // 服务器的可选列表，可以为空，为空时会自动获取
	ServerProxy       string           // 代理服务器地址(https)请求时，可以为空，为空时不使用代理
	GatewayProxy      string           // 代理网关服务器地址(wss)请求时，可以为空，为空时不使用代理
	GameProxy         string           // 代理游戏服务器地址(wss)请求时，可以为空，为空时不使用代理
	ReconnectInterval time.Duration    // 重连间隔时间
	ReconnectNumber   int              // 重连次数，当重连次数达到该值时, ws 连接不再尝试重连
}

type ConfigOption func(*Config)

func WithServerAddressList(serverAddressList []*ServerAddress) ConfigOption {
	return func(config *Config) {
		if len(serverAddressList) == 0 {
			logger.Error("serverAddressList is empty.")
			return
		}
		config.ServerAddressList = serverAddressList
	}
}

func WithServerProxy(proxyAddress string) ConfigOption {
	return func(config *Config) {
		config.ServerProxy = proxyAddress
	}
}

func WithGatewayProxy(proxyAddress string) ConfigOption {
	return func(config *Config) {
		config.GatewayProxy = proxyAddress
	}
}

func WithGameProxy(proxyAddress string) ConfigOption {
	return func(config *Config) {
		config.GameProxy = proxyAddress
	}
}

func WithReconnect(number int, interval time.Duration) ConfigOption {
	return func(config *Config) {
		config.ReconnectNumber = number
		config.ReconnectInterval = interval
	}
}

// Majsoul majsoul wsClient
type Majsoul struct {
	message.LobbyClient           // message.LobbyClient 更多时候在大厅时调用的是该接口
	LobbyConn           *wsClient // lobbyConn 是 message.LobbyClient 使用的连接

	message.FastTestClient           // message.FastTestClient 场景处于游戏桌面时调用该接口
	FastTestConn           *wsClient // fastTestConn 是 message.FastTestClient 使用的连接

	Implement Implement // 使得程序可以以多态的方式调用 message.LobbyClient 或 message.FastTestClient 的接口
	UUID      string

	ServerAddress *ServerAddress
	Request       *request // 用于直接向http(s)请求
	Version       *Version // 初始化时获取的版本信息

	Config   *Config
	Account  *message.Account     // 该字段应在登录成功后访问
	GameInfo *message.ResAuthGame // 该字段应在进入游戏桌面后访问
}

// New 初始化 majsoul
// ctx context.Context
// config *Config 这个入参可以为空，为空时会使用默认值，这个类型每一个字段都可以为空
func New(ctx context.Context, configOptions ...ConfigOption) (majsoul *Majsoul, err error) {
	config := &Config{}

	for _, configOption := range configOptions {
		configOption(config)
	}

	majsoul = &Majsoul{
		Config: config,
		UUID:   uuid(),
	}

	if len(config.ServerAddressList) == 0 {
		config.ServerAddressList = ServerAddressList
	}

	err = majsoul.tryNew(ctx)
	if err != nil {
		return nil, err
	}

	err = majsoul.initVersion(ctx)
	if err != nil {
		return nil, err
	}

	go majsoul.heatbeat(ctx)
	go majsoul.receiveConn(ctx)
	return majsoul, nil
}

func (majsoul *Majsoul) tryNew(ctx context.Context) (err error) {
	for _, serverAddress := range majsoul.Config.ServerAddressList {
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
			continue
		}
		client := newWsClient(&wsConfig{
			ConnAddress:       serverAddress.GatewayAddress,
			ProxyAddress:      majsoul.Config.GatewayProxy,
			RequestHeaders:    header,
			ReconnectInterval: majsoul.Config.ReconnectInterval,
			ReconnectNumber:   majsoul.Config.ReconnectNumber,
		})
		err = client.Connect(ctx)
		if err != nil {
			continue
		}
		majsoul.ServerAddress = serverAddress
		majsoul.Request = r
		majsoul.LobbyConn = client
		majsoul.LobbyClient = message.NewLobbyClient(client)
		majsoul.Implement = majsoul
		return nil
	}
	return ErrorNoServerAvailable
}

func (majsoul *Majsoul) initVersion(ctx context.Context) (err error) {
	majsoul.Version, err = majsoul.version()
	return
}

func (majsoul *Majsoul) ConnGame(ctx context.Context) (err error) {
	connUrl, err := url.Parse(majsoul.ServerAddress.GameAddress)
	if err != nil {
		logger.Error("failed to parse GameAddress: ", zap.String("GameAddress", majsoul.ServerAddress.GameAddress), zap.Error(err))
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
		ConnAddress:       majsoul.ServerAddress.GameAddress,
		ProxyAddress:      majsoul.Config.GameProxy,
		RequestHeaders:    header,
		ReconnectInterval: majsoul.Config.ReconnectInterval,
		ReconnectNumber:   majsoul.Config.ReconnectNumber,
	})
	err = clinet.Connect(ctx)
	if err != nil {
		logger.Error("failed to connect to GameServer: ", zap.String("GameAddress", majsoul.ServerAddress.GameAddress), zap.Error(err))
		return
	}

	majsoul.FastTestConn = clinet
	majsoul.FastTestClient = message.NewFastTestClient(majsoul.FastTestConn)
	go majsoul.receiveGame(ctx)
	return
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
			if majsoul.FastTestConn != nil {
				continue
			}
			_, err := majsoul.Heatbeat(ctx, &message.ReqHeatBeat{})
			if err != nil {
				logger.Error("Majsoul.heatbeat error:", zap.Error(err))
				return
			}
		case <-t2.C:
			if majsoul.FastTestConn == nil {
				continue
			}
			_, err := majsoul.CheckNetworkDelay(ctx, &message.ReqCommon{})
			if err != nil {
				logger.Error("Majsoul.checkNetworkDelay error:", zap.Error(err))
				return
			}
		}
	}
}

func (majsoul *Majsoul) receiveConn(ctx context.Context) {
	for {
		if majsoul.LobbyConn == nil {
			logger.Debug("lobbyConn lost")
			return
		}
		select {
		case <-ctx.Done():
			return
		case data := <-majsoul.LobbyConn.Receive():
			majsoul.handleNotify(ctx, data)
		}
	}
}

func (majsoul *Majsoul) receiveGame(ctx context.Context) {
	for {
		if majsoul.FastTestConn == nil {
			logger.Debug("fastTestConn lost")
			return
		}
		select {
		case <-ctx.Done():
			return
		case data := <-majsoul.FastTestConn.Receive():
			majsoul.handleNotify(ctx, data)
		}
	}
}

func (majsoul *Majsoul) handleNotify(ctx context.Context, data proto.Message) {
	if majsoul.Implement == nil {
		logger.Panic("majsoul.Implement is null")
	}
	switch notify := data.(type) {
	case *message.NotifyCaptcha:
		majsoul.NotifyCaptcha(ctx, notify)
		majsoul.Implement.NotifyCaptcha(ctx, notify)
	case *message.NotifyRoomGameStart:
		majsoul.NotifyRoomGameStart(ctx, notify)
		majsoul.Implement.NotifyRoomGameStart(ctx, notify)
	case *message.NotifyMatchGameStart:
		majsoul.NotifyMatchGameStart(ctx, notify)
		majsoul.Implement.NotifyMatchGameStart(ctx, notify)
	case *message.NotifyRoomPlayerReady:
		majsoul.NotifyRoomPlayerReady(ctx, notify)
		majsoul.Implement.NotifyRoomPlayerReady(ctx, notify)
	case *message.NotifyRoomPlayerDressing:
		majsoul.NotifyRoomPlayerDressing(ctx, notify)
		majsoul.Implement.NotifyRoomPlayerDressing(ctx, notify)
	case *message.NotifyRoomPlayerUpdate:
		majsoul.NotifyRoomPlayerUpdate(ctx, notify)
		majsoul.Implement.NotifyRoomPlayerUpdate(ctx, notify)
	case *message.NotifyRoomKickOut:
		majsoul.NotifyRoomKickOut(ctx, notify)
		majsoul.Implement.NotifyRoomKickOut(ctx, notify)
	case *message.NotifyFriendStateChange:
		majsoul.NotifyFriendStateChange(ctx, notify)
		majsoul.Implement.NotifyFriendStateChange(ctx, notify)
	case *message.NotifyFriendViewChange:
		majsoul.NotifyFriendViewChange(ctx, notify)
		majsoul.Implement.NotifyFriendViewChange(ctx, notify)
	case *message.NotifyFriendChange:
		majsoul.NotifyFriendChange(ctx, notify)
		majsoul.Implement.NotifyFriendChange(ctx, notify)
	case *message.NotifyNewFriendApply:
		majsoul.NotifyNewFriendApply(ctx, notify)
		majsoul.Implement.NotifyNewFriendApply(ctx, notify)
	case *message.NotifyClientMessage:
		majsoul.NotifyClientMessage(ctx, notify)
		majsoul.Implement.NotifyClientMessage(ctx, notify)
	case *message.NotifyAccountUpdate:
		majsoul.NotifyAccountUpdate(ctx, notify)
		majsoul.Implement.NotifyAccountUpdate(ctx, notify)
	case *message.NotifyAnotherLogin:
		majsoul.NotifyAnotherLogin(ctx, notify)
		majsoul.Implement.NotifyAnotherLogin(ctx, notify)
	case *message.NotifyAccountLogout:
		majsoul.NotifyAccountLogout(ctx, notify)
		majsoul.Implement.NotifyAccountLogout(ctx, notify)
	case *message.NotifyAnnouncementUpdate:
		majsoul.NotifyAnnouncementUpdate(ctx, notify)
		majsoul.Implement.NotifyAnnouncementUpdate(ctx, notify)
	case *message.NotifyNewMail:
		majsoul.NotifyNewMail(ctx, notify)
		majsoul.Implement.NotifyNewMail(ctx, notify)
	case *message.NotifyDeleteMail:
		majsoul.NotifyDeleteMail(ctx, notify)
		majsoul.Implement.NotifyDeleteMail(ctx, notify)
	case *message.NotifyReviveCoinUpdate:
		majsoul.NotifyReviveCoinUpdate(ctx, notify)
		majsoul.Implement.NotifyReviveCoinUpdate(ctx, notify)
	case *message.NotifyDailyTaskUpdate:
		majsoul.NotifyDailyTaskUpdate(ctx, notify)
		majsoul.Implement.NotifyDailyTaskUpdate(ctx, notify)
	case *message.NotifyActivityTaskUpdate:
		majsoul.NotifyActivityTaskUpdate(ctx, notify)
		majsoul.Implement.NotifyActivityTaskUpdate(ctx, notify)
	case *message.NotifyActivityPeriodTaskUpdate:
		majsoul.NotifyActivityPeriodTaskUpdate(ctx, notify)
		majsoul.Implement.NotifyActivityPeriodTaskUpdate(ctx, notify)
	case *message.NotifyAccountRandomTaskUpdate:
		majsoul.NotifyAccountRandomTaskUpdate(ctx, notify)
		majsoul.Implement.NotifyAccountRandomTaskUpdate(ctx, notify)
	case *message.NotifyActivitySegmentTaskUpdate:
		majsoul.NotifyActivitySegmentTaskUpdate(ctx, notify)
		majsoul.Implement.NotifyActivitySegmentTaskUpdate(ctx, notify)
	case *message.NotifyActivityUpdate:
		majsoul.NotifyActivityUpdate(ctx, notify)
		majsoul.Implement.NotifyActivityUpdate(ctx, notify)
	case *message.NotifyAccountChallengeTaskUpdate:
		majsoul.NotifyAccountChallengeTaskUpdate(ctx, notify)
		majsoul.Implement.NotifyAccountChallengeTaskUpdate(ctx, notify)
	case *message.NotifyNewComment:
		majsoul.NotifyNewComment(ctx, notify)
		majsoul.Implement.NotifyNewComment(ctx, notify)
	case *message.NotifyRollingNotice:
		majsoul.NotifyRollingNotice(ctx, notify)
		majsoul.Implement.NotifyRollingNotice(ctx, notify)
	case *message.NotifyGiftSendRefresh:
		majsoul.NotifyGiftSendRefresh(ctx, notify)
		majsoul.Implement.NotifyGiftSendRefresh(ctx, notify)
	case *message.NotifyShopUpdate:
		majsoul.NotifyShopUpdate(ctx, notify)
		majsoul.Implement.NotifyShopUpdate(ctx, notify)
	case *message.NotifyVipLevelChange:
		majsoul.NotifyVipLevelChange(ctx, notify)
		majsoul.Implement.NotifyVipLevelChange(ctx, notify)
	case *message.NotifyServerSetting:
		majsoul.NotifyServerSetting(ctx, notify)
		majsoul.Implement.NotifyServerSetting(ctx, notify)
	case *message.NotifyPayResult:
		majsoul.NotifyPayResult(ctx, notify)
		majsoul.Implement.NotifyPayResult(ctx, notify)
	case *message.NotifyCustomContestAccountMsg:
		majsoul.NotifyCustomContestAccountMsg(ctx, notify)
		majsoul.Implement.NotifyCustomContestAccountMsg(ctx, notify)
	case *message.NotifyCustomContestSystemMsg:
		majsoul.NotifyCustomContestSystemMsg(ctx, notify)
		majsoul.Implement.NotifyCustomContestSystemMsg(ctx, notify)
	case *message.NotifyMatchTimeout:
		majsoul.NotifyMatchTimeout(ctx, notify)
		majsoul.Implement.NotifyMatchTimeout(ctx, notify)
	case *message.NotifyCustomContestState:
		majsoul.NotifyCustomContestState(ctx, notify)
		majsoul.Implement.NotifyCustomContestState(ctx, notify)
	case *message.NotifyActivityChange:
		majsoul.NotifyActivityChange(ctx, notify)
		majsoul.Implement.NotifyActivityChange(ctx, notify)
	case *message.NotifyAFKResult:
		majsoul.NotifyAFKResult(ctx, notify)
		majsoul.Implement.NotifyAFKResult(ctx, notify)
	case *message.NotifyGameFinishRewardV2:
		majsoul.NotifyGameFinishRewardV2(ctx, notify)
		majsoul.Implement.NotifyGameFinishRewardV2(ctx, notify)
	case *message.NotifyActivityRewardV2:
		majsoul.NotifyActivityRewardV2(ctx, notify)
		majsoul.Implement.NotifyActivityRewardV2(ctx, notify)
	case *message.NotifyActivityPointV2:
		majsoul.NotifyActivityPointV2(ctx, notify)
		majsoul.Implement.NotifyActivityPointV2(ctx, notify)
	case *message.NotifyLeaderboardPointV2:
		majsoul.NotifyLeaderboardPointV2(ctx, notify)
		majsoul.Implement.NotifyLeaderboardPointV2(ctx, notify)
	case *message.NotifyNewGame:
		majsoul.NotifyNewGame(ctx, notify)
		majsoul.Implement.NotifyNewGame(ctx, notify)
	case *message.NotifyPlayerLoadGameReady:
		majsoul.NotifyPlayerLoadGameReady(ctx, notify)
		majsoul.Implement.NotifyPlayerLoadGameReady(ctx, notify)
	case *message.NotifyGameBroadcast:
		majsoul.NotifyGameBroadcast(ctx, notify)
		majsoul.Implement.NotifyGameBroadcast(ctx, notify)
	case *message.NotifyGameEndResult:
		majsoul.NotifyGameEndResult(ctx, notify)
		majsoul.Implement.NotifyGameEndResult(ctx, notify)
	case *message.NotifyGameTerminate:
		majsoul.NotifyGameTerminate(ctx, notify)
		majsoul.Implement.NotifyGameTerminate(ctx, notify)
	case *message.NotifyPlayerConnectionState:
		majsoul.NotifyPlayerConnectionState(ctx, notify)
		majsoul.Implement.NotifyPlayerConnectionState(ctx, notify)
	case *message.NotifyAccountLevelChange:
		majsoul.NotifyAccountLevelChange(ctx, notify)
		majsoul.Implement.NotifyAccountLevelChange(ctx, notify)
	case *message.NotifyGameFinishReward:
		majsoul.NotifyGameFinishReward(ctx, notify)
		majsoul.Implement.NotifyGameFinishReward(ctx, notify)
	case *message.NotifyActivityReward:
		majsoul.NotifyActivityReward(ctx, notify)
		majsoul.Implement.NotifyActivityReward(ctx, notify)
	case *message.NotifyActivityPoint:
		majsoul.NotifyActivityPoint(ctx, notify)
		majsoul.Implement.NotifyActivityPoint(ctx, notify)
	case *message.NotifyLeaderboardPoint:
		majsoul.NotifyLeaderboardPoint(ctx, notify)
		majsoul.Implement.NotifyLeaderboardPoint(ctx, notify)
	case *message.NotifyGamePause:
		majsoul.NotifyGamePause(ctx, notify)
		majsoul.Implement.NotifyGamePause(ctx, notify)
	case *message.NotifyEndGameVote:
		majsoul.NotifyEndGameVote(ctx, notify)
		majsoul.Implement.NotifyEndGameVote(ctx, notify)
	case *message.NotifyObserveData:
		majsoul.NotifyObserveData(ctx, notify)
		majsoul.Implement.NotifyObserveData(ctx, notify)
	case *message.NotifyRoomPlayerReady_AccountReadyState:
		majsoul.NotifyRoomPlayerReady_AccountReadyState(ctx, notify)
		majsoul.Implement.NotifyRoomPlayerReady_AccountReadyState(ctx, notify)
	case *message.NotifyRoomPlayerDressing_AccountDressingState:
		majsoul.NotifyRoomPlayerDressing_AccountDressingState(ctx, notify)
		majsoul.Implement.NotifyRoomPlayerDressing_AccountDressingState(ctx, notify)
	case *message.NotifyAnnouncementUpdate_AnnouncementUpdate:
		majsoul.NotifyAnnouncementUpdate_AnnouncementUpdate(ctx, notify)
		majsoul.Implement.NotifyAnnouncementUpdate_AnnouncementUpdate(ctx, notify)
	case *message.NotifyActivityUpdate_FeedActivityData:
		majsoul.NotifyActivityUpdate_FeedActivityData(ctx, notify)
		majsoul.Implement.NotifyActivityUpdate_FeedActivityData(ctx, notify)
	case *message.NotifyActivityUpdate_FeedActivityData_CountWithTimeData:
		majsoul.NotifyActivityUpdate_FeedActivityData_CountWithTimeData(ctx, notify)
		majsoul.Implement.NotifyActivityUpdate_FeedActivityData_CountWithTimeData(ctx, notify)
	case *message.NotifyActivityUpdate_FeedActivityData_GiftBoxData:
		majsoul.NotifyActivityUpdate_FeedActivityData_GiftBoxData(ctx, notify)
		majsoul.Implement.NotifyActivityUpdate_FeedActivityData_GiftBoxData(ctx, notify)
	case *message.NotifyPayResult_ResourceModify:
		majsoul.NotifyPayResult_ResourceModify(ctx, notify)
		majsoul.Implement.NotifyPayResult_ResourceModify(ctx, notify)
	case *message.NotifyGameFinishRewardV2_LevelChange:
		majsoul.NotifyGameFinishRewardV2_LevelChange(ctx, notify)
		majsoul.Implement.NotifyGameFinishRewardV2_LevelChange(ctx, notify)
	case *message.NotifyGameFinishRewardV2_MatchChest:
		majsoul.NotifyGameFinishRewardV2_MatchChest(ctx, notify)
		majsoul.Implement.NotifyGameFinishRewardV2_MatchChest(ctx, notify)
	case *message.NotifyGameFinishRewardV2_MainCharacter:
		majsoul.NotifyGameFinishRewardV2_MainCharacter(ctx, notify)
		majsoul.Implement.NotifyGameFinishRewardV2_MainCharacter(ctx, notify)
	case *message.NotifyGameFinishRewardV2_CharacterGift:
		majsoul.NotifyGameFinishRewardV2_CharacterGift(ctx, notify)
		majsoul.Implement.NotifyGameFinishRewardV2_CharacterGift(ctx, notify)
	case *message.NotifyActivityRewardV2_ActivityReward:
		majsoul.NotifyActivityRewardV2_ActivityReward(ctx, notify)
		majsoul.Implement.NotifyActivityRewardV2_ActivityReward(ctx, notify)
	case *message.NotifyActivityPointV2_ActivityPoint:
		majsoul.NotifyActivityPointV2_ActivityPoint(ctx, notify)
		majsoul.Implement.NotifyActivityPointV2_ActivityPoint(ctx, notify)
	case *message.NotifyLeaderboardPointV2_LeaderboardPoint:
		majsoul.NotifyLeaderboardPointV2_LeaderboardPoint(ctx, notify)
		majsoul.Implement.NotifyLeaderboardPointV2_LeaderboardPoint(ctx, notify)
	case *message.NotifyGameFinishReward_LevelChange:
		majsoul.NotifyGameFinishReward_LevelChange(ctx, notify)
		majsoul.Implement.NotifyGameFinishReward_LevelChange(ctx, notify)
	case *message.NotifyGameFinishReward_MatchChest:
		majsoul.NotifyGameFinishReward_MatchChest(ctx, notify)
		majsoul.Implement.NotifyGameFinishReward_MatchChest(ctx, notify)
	case *message.NotifyGameFinishReward_MainCharacter:
		majsoul.NotifyGameFinishReward_MainCharacter(ctx, notify)
		majsoul.Implement.NotifyGameFinishReward_MainCharacter(ctx, notify)
	case *message.NotifyGameFinishReward_CharacterGift:
		majsoul.NotifyGameFinishReward_CharacterGift(ctx, notify)
		majsoul.Implement.NotifyGameFinishReward_CharacterGift(ctx, notify)
	case *message.NotifyActivityReward_ActivityReward:
		majsoul.NotifyActivityReward_ActivityReward(ctx, notify)
		majsoul.Implement.NotifyActivityReward_ActivityReward(ctx, notify)
	case *message.NotifyActivityPoint_ActivityPoint:
		majsoul.NotifyActivityPoint_ActivityPoint(ctx, notify)
		majsoul.Implement.NotifyActivityPoint_ActivityPoint(ctx, notify)
	case *message.NotifyLeaderboardPoint_LeaderboardPoint:
		majsoul.NotifyLeaderboardPoint_LeaderboardPoint(ctx, notify)
		majsoul.Implement.NotifyLeaderboardPoint_LeaderboardPoint(ctx, notify)
	case *message.NotifyEndGameVote_VoteResult:
		majsoul.NotifyEndGameVote_VoteResult(ctx, notify)
		majsoul.Implement.NotifyEndGameVote_VoteResult(ctx, notify)
	case *message.PlayerLeaving:
		majsoul.PlayerLeaving(ctx, notify)
		majsoul.Implement.PlayerLeaving(ctx, notify)
	case *message.ActionPrototype:
		majsoul.ActionPrototype(ctx, notify)
		majsoul.Implement.ActionPrototype(ctx, notify)
	default:
		logger.Info("unknown notify type", zap.Reflect("notify", notify))
	}
}

func uuid() string {
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

// message.LobbyClient

func (majsoul *Majsoul) OnReconnect(callbreak func(ctx context.Context)) {
	majsoul.LobbyConn.ReconnectHandler = callbreak
}

func (majsoul *Majsoul) Login(ctx context.Context, account, password string) (*message.ResLogin, error) {
	var t uint32
	if !strings.Contains(account, "@") {
		t = 1
	}
	loginRes, err := majsoul.LobbyClient.Login(ctx, &message.ReqLogin{
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
	if loginRes.Error == nil {
		majsoul.Account = loginRes.Account
	}
	return loginRes, nil
}

// message.FastTestClient
