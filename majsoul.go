// Package majsoul https://game.maj-soul.com/1/
package majsoul

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/constellation39/majsoul/message"
	"github.com/constellation39/majsoul/utils"
	"github.com/golang/protobuf/proto"
)

const (
	MsgTypeNotify   uint8 = 1
	MsgTypeRequest  uint8 = 2
	MsgTypeResponse uint8 = 3

	Discard = 1
	Chi     = 2
	Pon     = 3
	AnKAN   = 4
	MinKan  = 5
	KaKan   = 6
	Riichi  = 7
	Tsumo   = 8
	Ron     = 9
	Kuku    = 10
	Kita    = 11
	Pass    = 12

	charSet  = "0123456789abcdefghijklmnopqrstuvwxyz"
	uuidFile = ".uuid"
)

type Implement interface {
	IFNotify // IFNotify 大厅通知下发
	IFAction // IFAction  游戏桌面内下发
}

// Majsoul majsoul client
type Majsoul struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	message.LobbyClient             // message.LobbyClient 更多时候在大厅时调用的是该接口
	LobbyConn           *ClientConn // LobbyConn 是 message.LobbyClient 使用的连接

	message.FastTestClient             // message.FastTestClient 场景处于游戏桌面时调用该接口
	FastTestConn           *ClientConn // FastTestConn 是 message.FastTestClient 使用的连接

	Implement     // 使得程序可以以多态的方式调用 message.LobbyClient 或 message.FastTestClient 的接口
	UUID          string
	ServerAddress *ServerAddress

	Request *utils.Request // 用于直接向http(s)请求
	Version *Version       // 初始化时获取的版本信息

	Account  *message.Account     // 该字段应在登录成功后访问
	GameInfo *message.ResAuthGame // 该字段应在进入游戏桌面后访问
}

func New() (*Majsoul, error) {
	ctx, cancel := context.WithCancel(context.Background())
	serverAddress, request, cConn, err := lookup(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	majsoul := &Majsoul{
		Ctx:           ctx,
		Cancel:        cancel,
		Request:       request,
		ServerAddress: serverAddress,
		LobbyClient:   message.NewLobbyClient(cConn),
		LobbyConn:     cConn,
	}
	majsoul.init()
	go majsoul.heatbeat()
	go majsoul.receiveConn()
	return majsoul, nil
}

func lookup(ctx context.Context) (*ServerAddress, *utils.Request, *ClientConn, error) {
	for _, serverAddress := range ServerAddressList {
		request := utils.NewRequest(serverAddress.ServerAddress)
		r := int(rand.Float32()*1000000000) + int(rand.Float32()*1000000000)
		_, err := request.Get(fmt.Sprintf("1/version.json?randv=%d", r))
		if err != nil {
			continue
		}
		cConn, err := NewClientConn(ctx, serverAddress.GatewayAddress)
		if err != nil {
			continue
		}
		return serverAddress, request, cConn, nil
	}
	return nil, nil, nil, fmt.Errorf("no servers were found that could be used")
}

func (majsoul *Majsoul) init() {
	var err error
	majsoul.Version, err = majsoul.version()

	if err != nil {
		log.Fatalf("Majsoul.init version error: %v \n", err)
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

func (majsoul *Majsoul) version() (*Version, error) {
	// var version_url = "version.json?randv="+Math.floor(Math.random() * 1000000000).toString()+Math.floor(Math.random() * 1000000000).toString()
	r := int(rand.Float32()*1000000000) + int(rand.Float32()*1000000000)
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

func (majsoul *Majsoul) heatbeat() {
	t3 := time.NewTicker(time.Second * 3)
	t2 := time.NewTicker(time.Second * 2)
loop:
	for {
		select {
		case <-t3.C:
			if majsoul.FastTestConn != nil {
				continue
			}
			_, err := majsoul.Heatbeat(majsoul.Ctx, &message.ReqHeatBeat{})
			if err != nil {
				fmt.Println("heatbeat error:", err)
				return
			}
		case <-t2.C:
			if majsoul.FastTestConn == nil {
				continue
			}
			_, err := majsoul.CheckNetworkDelay(majsoul.Ctx, &message.ReqCommon{})
			if err != nil {
				fmt.Println("checkNetworkDelay error:", err)
				return
			}
		case <-majsoul.Ctx.Done():
			break loop
		}
	}
}

func (majsoul *Majsoul) receiveConn() {
	for data := range majsoul.LobbyConn.Receive() {
		majsoul.handleNotify(data)
	}
}

func (majsoul *Majsoul) receiveGame() {
	for data := range majsoul.FastTestConn.Receive() {
		majsoul.handleNotify(data)
	}
}

func (majsoul *Majsoul) handleNotify(data proto.Message) {
	switch notify := data.(type) {
	case *message.NotifyCaptcha:
		majsoul.Implement.NotifyCaptcha(notify)
	case *message.NotifyRoomGameStart:
		majsoul.Implement.NotifyRoomGameStart(notify)
	case *message.NotifyMatchGameStart:
		majsoul.Implement.NotifyMatchGameStart(notify)
	case *message.NotifyRoomPlayerReady:
		majsoul.Implement.NotifyRoomPlayerReady(notify)
	case *message.NotifyRoomPlayerDressing:
		majsoul.Implement.NotifyRoomPlayerDressing(notify)
	case *message.NotifyRoomPlayerUpdate:
		majsoul.Implement.NotifyRoomPlayerUpdate(notify)
	case *message.NotifyRoomKickOut:
		majsoul.Implement.NotifyRoomKickOut(notify)
	case *message.NotifyFriendStateChange:
		majsoul.Implement.NotifyFriendStateChange(notify)
	case *message.NotifyFriendViewChange:
		majsoul.Implement.NotifyFriendViewChange(notify)
	case *message.NotifyFriendChange:
		majsoul.Implement.NotifyFriendChange(notify)
	case *message.NotifyNewFriendApply:
		majsoul.Implement.NotifyNewFriendApply(notify)
	case *message.NotifyClientMessage:
		majsoul.Implement.NotifyClientMessage(notify)
	case *message.NotifyAccountUpdate:
		majsoul.Implement.NotifyAccountUpdate(notify)
	case *message.NotifyAnotherLogin:
		majsoul.Implement.NotifyAnotherLogin(notify)
	case *message.NotifyAccountLogout:
		majsoul.Implement.NotifyAccountLogout(notify)
	case *message.NotifyAnnouncementUpdate:
		majsoul.Implement.NotifyAnnouncementUpdate(notify)
	case *message.NotifyNewMail:
		majsoul.Implement.NotifyNewMail(notify)
	case *message.NotifyDeleteMail:
		majsoul.Implement.NotifyDeleteMail(notify)
	case *message.NotifyReviveCoinUpdate:
		majsoul.Implement.NotifyReviveCoinUpdate(notify)
	case *message.NotifyDailyTaskUpdate:
		majsoul.Implement.NotifyDailyTaskUpdate(notify)
	case *message.NotifyActivityTaskUpdate:
		majsoul.Implement.NotifyActivityTaskUpdate(notify)
	case *message.NotifyActivityPeriodTaskUpdate:
		majsoul.Implement.NotifyActivityPeriodTaskUpdate(notify)
	case *message.NotifyAccountRandomTaskUpdate:
		majsoul.Implement.NotifyAccountRandomTaskUpdate(notify)
	case *message.NotifyActivitySegmentTaskUpdate:
		majsoul.Implement.NotifyActivitySegmentTaskUpdate(notify)
	case *message.NotifyActivityUpdate:
		majsoul.Implement.NotifyActivityUpdate(notify)
	case *message.NotifyAccountChallengeTaskUpdate:
		majsoul.Implement.NotifyAccountChallengeTaskUpdate(notify)
	case *message.NotifyNewComment:
		majsoul.Implement.NotifyNewComment(notify)
	case *message.NotifyRollingNotice:
		majsoul.Implement.NotifyRollingNotice(notify)
	case *message.NotifyGiftSendRefresh:
		majsoul.Implement.NotifyGiftSendRefresh(notify)
	case *message.NotifyShopUpdate:
		majsoul.Implement.NotifyShopUpdate(notify)
	case *message.NotifyVipLevelChange:
		majsoul.Implement.NotifyVipLevelChange(notify)
	case *message.NotifyServerSetting:
		majsoul.Implement.NotifyServerSetting(notify)
	case *message.NotifyPayResult:
		majsoul.Implement.NotifyPayResult(notify)
	case *message.NotifyCustomContestAccountMsg:
		majsoul.Implement.NotifyCustomContestAccountMsg(notify)
	case *message.NotifyCustomContestSystemMsg:
		majsoul.Implement.NotifyCustomContestSystemMsg(notify)
	case *message.NotifyMatchTimeout:
		majsoul.Implement.NotifyMatchTimeout(notify)
	case *message.NotifyCustomContestState:
		majsoul.Implement.NotifyCustomContestState(notify)
	case *message.NotifyActivityChange:
		majsoul.Implement.NotifyActivityChange(notify)
	case *message.NotifyAFKResult:
		majsoul.Implement.NotifyAFKResult(notify)
	case *message.NotifyGameFinishRewardV2:
		majsoul.Implement.NotifyGameFinishRewardV2(notify)
	case *message.NotifyActivityRewardV2:
		majsoul.Implement.NotifyActivityRewardV2(notify)
	case *message.NotifyActivityPointV2:
		majsoul.Implement.NotifyActivityPointV2(notify)
	case *message.NotifyLeaderboardPointV2:
		majsoul.Implement.NotifyLeaderboardPointV2(notify)
	case *message.NotifyNewGame:
		majsoul.Implement.NotifyNewGame(notify)
	case *message.NotifyPlayerLoadGameReady:
		majsoul.Implement.NotifyPlayerLoadGameReady(notify)
	case *message.NotifyGameBroadcast:
		majsoul.Implement.NotifyGameBroadcast(notify)
	case *message.NotifyGameEndResult:
		majsoul.Implement.NotifyGameEndResult(notify)
	case *message.NotifyGameTerminate:
		majsoul.Implement.NotifyGameTerminate(notify)
	case *message.NotifyPlayerConnectionState:
		majsoul.Implement.NotifyPlayerConnectionState(notify)
	case *message.NotifyAccountLevelChange:
		majsoul.Implement.NotifyAccountLevelChange(notify)
	case *message.NotifyGameFinishReward:
		majsoul.Implement.NotifyGameFinishReward(notify)
	case *message.NotifyActivityReward:
		majsoul.Implement.NotifyActivityReward(notify)
	case *message.NotifyActivityPoint:
		majsoul.Implement.NotifyActivityPoint(notify)
	case *message.NotifyLeaderboardPoint:
		majsoul.Implement.NotifyLeaderboardPoint(notify)
	case *message.NotifyGamePause:
		majsoul.Implement.NotifyGamePause(notify)
	case *message.NotifyEndGameVote:
		majsoul.Implement.NotifyEndGameVote(notify)
	case *message.NotifyObserveData:
		majsoul.Implement.NotifyObserveData(notify)
	case *message.NotifyRoomPlayerReady_AccountReadyState:
		majsoul.Implement.NotifyRoomPlayerReady_AccountReadyState(notify)
	case *message.NotifyRoomPlayerDressing_AccountDressingState:
		majsoul.Implement.NotifyRoomPlayerDressing_AccountDressingState(notify)
	case *message.NotifyAnnouncementUpdate_AnnouncementUpdate:
		majsoul.Implement.NotifyAnnouncementUpdate_AnnouncementUpdate(notify)
	case *message.NotifyActivityUpdate_FeedActivityData:
		majsoul.Implement.NotifyActivityUpdate_FeedActivityData(notify)
	case *message.NotifyActivityUpdate_FeedActivityData_CountWithTimeData:
		majsoul.Implement.NotifyActivityUpdate_FeedActivityData_CountWithTimeData(notify)
	case *message.NotifyActivityUpdate_FeedActivityData_GiftBoxData:
		majsoul.Implement.NotifyActivityUpdate_FeedActivityData_GiftBoxData(notify)
	case *message.NotifyPayResult_ResourceModify:
		majsoul.Implement.NotifyPayResult_ResourceModify(notify)
	case *message.NotifyGameFinishRewardV2_LevelChange:
		majsoul.Implement.NotifyGameFinishRewardV2_LevelChange(notify)
	case *message.NotifyGameFinishRewardV2_MatchChest:
		majsoul.Implement.NotifyGameFinishRewardV2_MatchChest(notify)
	case *message.NotifyGameFinishRewardV2_MainCharacter:
		majsoul.Implement.NotifyGameFinishRewardV2_MainCharacter(notify)
	case *message.NotifyGameFinishRewardV2_CharacterGift:
		majsoul.Implement.NotifyGameFinishRewardV2_CharacterGift(notify)
	case *message.NotifyActivityRewardV2_ActivityReward:
		majsoul.Implement.NotifyActivityRewardV2_ActivityReward(notify)
	case *message.NotifyActivityPointV2_ActivityPoint:
		majsoul.Implement.NotifyActivityPointV2_ActivityPoint(notify)
	case *message.NotifyLeaderboardPointV2_LeaderboardPoint:
		majsoul.Implement.NotifyLeaderboardPointV2_LeaderboardPoint(notify)
	case *message.NotifyGameFinishReward_LevelChange:
		majsoul.Implement.NotifyGameFinishReward_LevelChange(notify)
	case *message.NotifyGameFinishReward_MatchChest:
		majsoul.Implement.NotifyGameFinishReward_MatchChest(notify)
	case *message.NotifyGameFinishReward_MainCharacter:
		majsoul.Implement.NotifyGameFinishReward_MainCharacter(notify)
	case *message.NotifyGameFinishReward_CharacterGift:
		majsoul.Implement.NotifyGameFinishReward_CharacterGift(notify)
	case *message.NotifyActivityReward_ActivityReward:
		majsoul.Implement.NotifyActivityReward_ActivityReward(notify)
	case *message.NotifyActivityPoint_ActivityPoint:
		majsoul.Implement.NotifyActivityPoint_ActivityPoint(notify)
	case *message.NotifyLeaderboardPoint_LeaderboardPoint:
		majsoul.Implement.NotifyLeaderboardPoint_LeaderboardPoint(notify)
	case *message.NotifyEndGameVote_VoteResult:
		majsoul.Implement.NotifyEndGameVote_VoteResult(notify)
	case *message.PlayerLeaving:
		majsoul.Implement.PlayerLeaving(notify)
	case *message.ActionPrototype:
		majsoul.Implement.ActionPrototype(notify)
	default:
		log.Printf("Majsoul.handleNotify unknown message type: %T \n", notify)
	}
}

// message.LobbyClient

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

func (majsoul *Majsoul) Login(account, password string) (*message.ResLogin, error) {
	if majsoul.UUID == "" {
		majsoul.UUID = uuid()
		err := utils.SaveFile(uuidFile, []byte(majsoul.UUID))
		if err != nil {
			return nil, err
		}
	}
	loginRes, err := majsoul.LobbyClient.Login(majsoul.Ctx, &message.ReqLogin{
		Account:   account,
		Password:  utils.Hash(password),
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
			ScreenWidth:    914,
			ScreenHeight:   1316,
		},
		RandomKey: majsoul.UUID,
		ClientVersion: &message.ClientVersionInfo{
			Resource: majsoul.Version.Version,
			Package:  "",
		},
		GenAccessToken:    true,
		CurrencyPlatforms: []uint32{2, 6, 8, 10, 11},
		// 电话1 邮箱0
		Type:                0,
		Version:             0,
		ClientVersionString: majsoul.Version.Web(),
	})
	if err != nil {
		return nil, err
	}
	majsoul.Account = loginRes.Account
	return loginRes, nil
}

// message.FastTestClient
