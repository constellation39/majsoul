package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"time"
)

type GameState struct {
	seat         uint32
	account      *message.Account     // 该字段应在登录成功后访问
	gameInfo     *message.ResAuthGame // 该字段应在进入游戏桌面后访问
	accessToken  string               // 验证身份时使用 的 token
	connectToken string               // 重连时使用的 token
	gameUuid     string               // 是否在游戏中
}

func (GameState) NotifyClientMessage(majSoul *majsoul.MajSoul, notifyClientMessage *message.NotifyClientMessage) {
	type DetailRule struct {
		TimeFixed    int  `json:"time_fixed"`
		TimeAdd      int  `json:"time_add"`
		DoraCount    int  `json:"dora_count"`
		Shiduan      int  `json:"shiduan"`
		InitPoint    int  `json:"init_point"`
		Fandian      int  `json@:"fandian"`
		Bianjietishi bool `json:"bianjietishi"`
		AiLevel      int  `json:"ai_level"`
		Fanfu        int  `json:"fanfu"`
		GuyiMode     int  `json:"guyi_mode"`
		OpenHand     int  `json:"open_hand"`
	}
	type Mode struct {
		Mode       int        `json:"mode"`
		Ai         bool       `json:"ai"`
		DetailRule DetailRule `json:"detail_rule"`
	}
	type InvitationRoom struct {
		RoomID    uint32 `json:"room_id"`
		Mode      Mode   `json:"mode"`
		Nickname  string `json:"nickname"`
		Verified  int    `json:"verified"`
		AccountID int    `json:"account_id"`
	}
	// 我们现在只处理 type == 1 , 也就是收到邀请的情况
	if notifyClientMessage.Type != 1 {
		logger.Info("notifyClientMessage.Type != -1", zap.Uint32("type", notifyClientMessage.Type))
		return
	}
	invitationRoom := new(InvitationRoom)
	err := json.Unmarshal([]byte(notifyClientMessage.Content), invitationRoom)
	if err != nil {
		logger.Error("Unmarshal", zap.Error(err))
		return
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		// 加入房间
		_, err = majSoul.LobbyClient.JoinRoom(ctx, &message.ReqJoinRoom{
			RoomId:              invitationRoom.RoomID,
			ClientVersionString: majSoul.Version.Web(),
		})
		if err != nil {
			logger.Error("JoinRoom", zap.Error(err))
			return
		}
	}

	time.Sleep(time.Second)

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		// 准备
		_, err = majSoul.LobbyClient.ReadyPlay(ctx, &message.ReqRoomReady{Ready: true})
		if err != nil {
			logger.Error("ReadyPlay", zap.Error(err))
			return
		}
	}
}

func (GameState) NotifyFriendViewChange(majSoul *majsoul.MajSoul, notifyFriendViewChange *message.NotifyFriendViewChange) {
	logger.Debug("", zap.Reflect("notifyFriendViewChange", notifyFriendViewChange))
}

// NotifyEndGameVote 有人发起投降
func (GameState) NotifyEndGameVote(majSoul *majsoul.MajSoul, notifyEndGameVote *message.NotifyEndGameVote) {
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.FastTestClient.VoteGameEnd(ctx, &message.ReqVoteGameEnd{Yes: true})
		if err != nil {
			logger.Error("VoteGameEnd", zap.Error(err))
		}
	}
}

// 从等待房间进入游戏时
func (gameState *GameState) NotifyRoomGameStart(majSoul *majsoul.MajSoul, notifyRoomGameStart *message.NotifyRoomGameStart) {

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := majSoul.ConnGame(ctx)
		if err != nil {
			panic(fmt.Sprintf("conn GameState server failed error %v", err))
		}
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		var err error
		gameState.gameInfo, err = majSoul.FastTestClient.AuthGame(ctx, &message.ReqAuthGame{
			AccountId: gameState.account.AccountId,
			Token:     gameState.connectToken,
			GameUuid:  gameState.gameUuid,
		})
		if err != nil {
			logger.Error("majsoul NotifyRoomGameStart AuthGame error: ", zap.Error(err))
			return
		}
	}
	gameState.connectToken = notifyRoomGameStart.ConnectToken
	gameState.gameUuid = notifyRoomGameStart.GameUuid
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.FastTestClient.EnterGame(ctx, &message.ReqCommon{})
		if err != nil {
			logger.Error("majsoul NotifyRoomGameStart EnterGame error:", zap.Error(err))
			return
		}
	}

	// 记录自己的座位号
	for i, uid := range gameState.gameInfo.SeatList {
		if uid == gameState.account.AccountId {
			gameState.seat = uint32(i)
			break
		}
	}
}

// ActionMJStart 游戏开始
func (GameState) ActionMJStart(majSoul *majsoul.MajSoul, actionMJStart *message.ActionMJStart) {
}

// ActionNewRound 回合开始
func (GameState) ActionNewRound(majSoul *majsoul.MajSoul, action *message.ActionNewRound) {
	// 如果是庄家
	if len(action.Tiles) != 14 {
		return
	}
	tile13 := action.Tiles[13]
	time.Sleep(time.Second * 3)
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.FastTestClient.InputOperation(ctx, &message.ReqSelfOperation{
			Type:    majsoul.ActionDiscard,
			Tile:    tile13,
			Moqie:   true,
			Timeuse: 1,
		})
		if err != nil {
			logger.Error("InputOperation failed", zap.Error(err))
		}
	}
}

// ActionDealTile 摸牌
func (gameState *GameState) ActionDealTile(majSoul *majsoul.MajSoul, action *message.ActionDealTile) {
	// 如果不是自己摸牌
	if action.Seat != gameState.seat {
		return
	}

	if len(action.Tile) == 0 {
		logger.Error("摸牌是空的")
		return
	}

	time.Sleep(time.Second * 3)
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.FastTestClient.InputOperation(ctx, &message.ReqSelfOperation{
			Type:    majsoul.ActionDiscard,
			Tile:    action.Tile,
			Moqie:   true,
			Timeuse: 1,
		})
		if err != nil {
			logger.Error("InputOperation failed", zap.Error(err))
		}
	}
}

// ActionDiscardTile 打牌
func (GameState) ActionDiscardTile(majSoul *majsoul.MajSoul, action *message.ActionDiscardTile) {
	if action.Operation != nil && len(action.Operation.OperationList) != 0 {
		for _, operation := range action.Operation.OperationList {
			switch operation.Type {
			case majsoul.ActionDiscard:
			case majsoul.ActionChi:
			case majsoul.ActionPon:
			case majsoul.ActionAnKAN:
			case majsoul.ActionMinKan:
			case majsoul.ActionKaKan:
			case majsoul.ActionRiichi:
			case majsoul.ActionTsumo:
			case majsoul.ActionRon:
			case majsoul.ActionKuku:
			case majsoul.ActionKita:
			case majsoul.ActionPass:
				if action.Operation != nil {
					{
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
						defer cancel()
						_, err := majSoul.FastTestClient.InputOperation(ctx, &message.ReqSelfOperation{
							CancelOperation: true,
							Timeuse:         1,
						})
						if err != nil {
							logger.Error("InputOperation failed", zap.Error(err))
						}
					}
				}
			}
		}
	}
}

// ActionChiPengGang 吃碰杠的通知
func (GameState) ActionChiPengGang(majSoul *majsoul.MajSoul, action *message.ActionChiPengGang) {
	switch action.Type {
	case majsoul.NotifyChi:
	case majsoul.NotifyPon:
	case majsoul.NotifyKan:
	}
}

// ActionAnGangAddGang 暗杠和加杠的通知
func (GameState) ActionAnGangAddGang(majSoul *majsoul.MajSoul, action *message.ActionAnGangAddGang) {
	switch action.Type {
	case majsoul.NotifyAnKan:
	case majsoul.NotifyKaKan:
	}
}

func (GameState) ActionHule(majSoul *majsoul.MajSoul, action *message.ActionHule) {
}

func (GameState) ActionLiuJu(majSoul *majsoul.MajSoul, action *message.ActionLiuJu) {
}

func (GameState) ActionNoTile(majSoul *majsoul.MajSoul, action *message.ActionNoTile) {
}

func main() {
	account, exists := os.LookupEnv("account")
	if !exists {
		panic("account is required.")
	}

	password, exists := os.LookupEnv("password")
	if !exists {
		panic("account is required.")
	}

	sync := logger.Init()
	defer sync()

	majSoul := majsoul.NewMajSoul(&majsoul.Config{ProxyAddress: ""})
	err := majSoul.LookupGateway(context.Background(), majsoul.ServerAddressList)
	if err != nil {
		panic(err)
	}

	logger.Debug("ServerAddress", zap.Reflect("ServerAddress", majSoul.ServerAddress))

	var gameState GameState

	majSoul.Handle(
		gameState.NotifyClientMessage,
		gameState.NotifyFriendViewChange,
		gameState.NotifyEndGameVote,
		gameState.NotifyRoomGameStart,
		gameState.ActionMJStart,
		gameState.ActionNewRound,
		gameState.ActionDealTile,
		gameState.ActionDiscardTile,
		gameState.ActionChiPengGang,
		gameState.ActionAnGangAddGang,
		gameState.ActionHule,
		gameState.ActionLiuJu,
		gameState.ActionNoTile,
	)

	{ // 登录
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		resLogin, err := majSoul.Login(ctx, account, password)
		if err != nil {
			panic(err)
		}
		if resLogin.Error == nil {
			gameState.account = resLogin.Account
			gameState.accessToken = resLogin.AccessToken
			if resLogin.GameInfo != nil {
				gameState.connectToken = resLogin.GameInfo.ConnectToken
				gameState.gameUuid = resLogin.GameInfo.GameUuid
			}
		}

		UpdateLoginInfo(majSoul)

		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			friendList, err := majSoul.LobbyClient.FetchFriendList(ctx, &message.ReqCommon{})
			if err != nil {
				return
			}
			logger.Debug("", zap.Reflect("friendList", friendList))
		}

		// 重连到正在进行对局的游戏中
		if resLogin.Account != nil && resLogin.Account.RoomId != 0 {
			{
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if err := majSoul.ConnGame(ctx); err != nil {
					logger.Error("client ConnGame error.", zap.Error(err))
				}
			}

			{
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				var err error
				gameState.gameInfo, err = majSoul.FastTestClient.AuthGame(ctx, &message.ReqAuthGame{
					AccountId: resLogin.Account.AccountId,
					Token:     resLogin.GameInfo.ConnectToken,
					GameUuid:  resLogin.GameInfo.GameUuid,
				})
				if err != nil {
					logger.Error("client AuthGame error.", zap.Error(err))
				}
			}

			for i, uid := range gameState.gameInfo.SeatList {
				if uid == resLogin.Account.AccountId {
					gameState.seat = uint32(i)
					break
				}
			}

			{
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if resSyncGame, err := majSoul.FastTestClient.SyncGame(ctx, &message.ReqSyncGame{RoundId: "-1"}); err != nil {
					logger.Error("majSoul SyncGame error.", zap.Error(err))
				} else {
					logger.Debug("majSoul SyncGame.", zap.Reflect("resSyncGame", resSyncGame))
				}
			}

			{
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if _, err := majSoul.FastTestClient.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
					logger.Error("majSoul FetchGamePlayerState error.", zap.Error(err))
				} else {
					logger.Debug("majSoul FetchGamePlayerState.")
				}
			}

			{
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if _, err := majSoul.FastTestClient.FinishSyncGame(ctx, &message.ReqCommon{}); err != nil {
					logger.Error("majSoul FinishSyncGame error.", zap.Error(err))
				} else {
					logger.Debug("majSoul FinishSyncGame.")
				}
			}

			{
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if _, err := majSoul.FastTestClient.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
					logger.Error("majSoul FetchGamePlayerState error.", zap.Error(err))
				} else {
					logger.Debug("majSoul FetchGamePlayerState.")
				}
			}
		}
	}

	majSoul.OnGatewayReconnect(func() {
		if len(gameState.accessToken) == 0 {
			panic(fmt.Sprintf(""))
		}
		var err error
		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			_, err = majSoul.LobbyClient.Oauth2Check(ctx, &message.ReqOauth2Check{AccessToken: gameState.accessToken})
			if err != nil {
				panic(fmt.Sprintf("gateway Oauth2Check error %v", err))
			}
		}
		var resLogin *message.ResLogin
		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			resLogin, err = majSoul.LobbyClient.Oauth2Login(ctx, &message.ReqOauth2Login{
				AccessToken: resLogin.AccessToken,
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
	})

	majSoul.OnGameReconnect(func() {
		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			var err error
			_, err = majSoul.FastTestClient.AuthGame(ctx, &message.ReqAuthGame{
				AccountId: gameState.account.AccountId,
				Token:     gameState.connectToken,
				GameUuid:  gameState.gameUuid,
			})
			if err != nil {
				logger.Error("majSoul AuthGame error.", zap.Error(err))
				return
			}
		}

		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			resSyncGame, err := majSoul.FastTestClient.SyncGame(ctx, &message.ReqSyncGame{RoundId: "-1"})
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
			if _, err := majSoul.FastTestClient.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
				logger.Error("majSoul FetchGamePlayerState error.", zap.Error(err))
				return
			} else {
				logger.Debug("majSoul FetchGamePlayerState.")
			}
		}

		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if _, err := majSoul.FastTestClient.FinishSyncGame(ctx, &message.ReqCommon{}); err != nil {
				logger.Error("majSoul FinishSyncGame error.", zap.Error(err))
				return
			} else {
				logger.Debug("majSoul FinishSyncGame.")
			}
		}

		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if _, err := majSoul.FastTestClient.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
				logger.Error("majSoul FetchGamePlayerState error.", zap.Error(err))
				return
			} else {
				logger.Debug("majSoul FetchGamePlayerState.")
			}
		}
	})

	println("Start.")
	select {}
}

func UpdateLoginInfo(majSoul *majsoul.MajSoul) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resFetchLastPrivacy, err := majSoul.LobbyClient.FetchLastPrivacy(ctx, &message.ReqFetchLastPrivacy{})
	if err != nil {
		logger.Error("majSoul FetchLastPrivacy error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchLastPrivacy.", zap.Reflect("resFetchLastPrivacy", resFetchLastPrivacy))

	resFetchServerTime, err := majSoul.LobbyClient.FetchServerTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchServerTime error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchServerTime.", zap.Reflect("resFetchServerTime", resFetchServerTime))

	resServerSettings, err := majSoul.LobbyClient.FetchServerSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchServerSettings error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchServerSettings.", zap.Reflect("resServerSettings", resServerSettings))

	resConnectionInfo, err := majSoul.LobbyClient.FetchConnectionInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchConnectionInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchConnectionInfo.", zap.Reflect("resConnectionInfo", resConnectionInfo))

	resClientValue, err := majSoul.LobbyClient.FetchClientValue(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchClientValue error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchClientValue.", zap.Reflect("resClientValue", resClientValue))

	resFriendList, err := majSoul.LobbyClient.FetchFriendList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchFriendList error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchFriendList.", zap.Reflect("resFriendList", resFriendList))

	resFriendApplyList, err := majSoul.LobbyClient.FetchFriendApplyList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchFriendApplyList error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchFriendApplyList.", zap.Reflect("resFriendApplyList", resFriendApplyList))

	resFetchrecentFriend, err := majSoul.LobbyClient.FetchRecentFriend(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchRecentFriend.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchRecentFriend.", zap.Reflect("resFetchrecentFriend", resFetchrecentFriend))

	resMailInfo, err := majSoul.LobbyClient.FetchMailInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchMailInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchMailInfo.", zap.Reflect("resMailInfo", resMailInfo))

	resDailyTask, err := majSoul.LobbyClient.FetchDailyTask(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchDailyTask error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchDailyTask.", zap.Reflect("resDailyTask", resDailyTask))

	resReviveCoinInfo, err := majSoul.LobbyClient.FetchReviveCoinInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchReviveCoinInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchReviveCoinInfo.", zap.Reflect("resReviveCoinInfo", resReviveCoinInfo))

	resTitleList, err := majSoul.LobbyClient.FetchTitleList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchTitleList error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchTitleList.", zap.Reflect("resTitleList", resTitleList))

	resBagInfo, err := majSoul.LobbyClient.FetchBagInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchBagInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchBagInfo.", zap.Reflect("resBagInfo", resBagInfo))

	resShopInfo, err := majSoul.LobbyClient.FetchShopInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchShopInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchShopInfo.", zap.Reflect("resShopInfo", resShopInfo))

	resFetchShopInterval, err := majSoul.LobbyClient.FetchShopInterval(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchShopInterval error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchShopInterval.", zap.Reflect("resFetchShopInterval", resFetchShopInterval))

	resActivityList, err := majSoul.LobbyClient.FetchActivityList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchActivityList error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchActivityList.", zap.Reflect("resActivityList", resActivityList))

	resAccountActivityData, err := majSoul.LobbyClient.FetchAccountActivityData(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchAccountActivityData error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchAccountActivityData.", zap.Reflect("resAccountActivityData", resAccountActivityData))

	resFetchActivityInterval, err := majSoul.LobbyClient.FetchActivityInterval(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchActivityInterval error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchActivityInterval.", zap.Reflect("resFetchActivityInterval", resFetchActivityInterval))

	resActivityBuff, err := majSoul.LobbyClient.FetchActivityBuff(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchActivityBuff error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchActivityBuff.", zap.Reflect("resActivityBuff", resActivityBuff))

	resVipReward, err := majSoul.LobbyClient.FetchVipReward(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchVipReward error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchVipReward.", zap.Reflect("resVipReward", resVipReward))

	resMonthTicketInfo, err := majSoul.LobbyClient.FetchMonthTicketInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchMonthTicketInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchMonthTicketInfo.", zap.Reflect("resMonthTicketInfo", resMonthTicketInfo))

	resAchievement, err := majSoul.LobbyClient.FetchAchievement(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchAchievement error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchAchievement.", zap.Reflect("resAchievement", resAchievement))

	resCommentSetting, err := majSoul.LobbyClient.FetchCommentSetting(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchCommentSetting error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchCommentSetting.", zap.Reflect("resCommentSetting", resCommentSetting))

	resAccountSettings, err := majSoul.LobbyClient.FetchAccountSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchAccountSettings error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchAccountSettings.", zap.Reflect("resAccountSettings", resAccountSettings))

	resModNicknameTime, err := majSoul.LobbyClient.FetchModNicknameTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchModNicknameTime error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchModNicknameTime.", zap.Reflect("resModNicknameTime", resModNicknameTime))

	resMisc, err := majSoul.LobbyClient.FetchMisc(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchMisc error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchMisc.", zap.Reflect("resMisc", resMisc))

	resAnnouncement, err := majSoul.LobbyClient.FetchAnnouncement(ctx, &message.ReqFetchAnnouncement{})
	if err != nil {
		logger.Error("majSoul FetchAnnouncement error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchAnnouncement.", zap.Reflect("resAnnouncement", resAnnouncement))

	// 写错了吧 req?
	reqRollingNotice, err := majSoul.LobbyClient.FetchRollingNotice(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul FetchRollingNotice error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul FetchRollingNotice.", zap.Reflect("reqRollingNotice", reqRollingNotice))

	resCommon, err := majSoul.LobbyClient.LoginSuccess(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majSoul LoginSuccess error.", zap.Error(err))
		return err
	}
	logger.Info("majSoul LoginSuccess.", zap.Reflect("resCommon", resCommon))

	return nil
}
