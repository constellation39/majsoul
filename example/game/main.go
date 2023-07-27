package main

import (
	"context"
	"fmt"
	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"time"
)

func ReConnGame(gameState *GameState, majSoul *majsoul.MajSoul, resLogin *message.ResLogin) {
	if resLogin.Account == nil || resLogin.Account.RoomId == 0 {
		return
	}
	{ // 尝试重连到游戏中
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := majSoul.ConnGame(ctx); err != nil {
			logger.Panic("client ConnGame error.", zap.Error(err))
		}
	}

	{ // 尝试验证身份
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		var err error
		gameState.gameInfo, err = majSoul.FastTestClient.AuthGame(ctx, &message.ReqAuthGame{
			AccountId: resLogin.Account.AccountId,
			Token:     resLogin.GameInfo.ConnectToken,
			GameUuid:  resLogin.GameInfo.GameUuid,
		})
		if err != nil {
			logger.Panic("client AuthGame error.", zap.Error(err))
		}
	}

	// 记录自己的座位号
	for i, uid := range gameState.gameInfo.SeatList {
		if uid == resLogin.Account.AccountId {
			gameState.seat = uint32(i)
			break
		}
	}

	{ // 尝试同步游戏数据
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if resSyncGame, err := majSoul.FastTestClient.SyncGame(ctx, &message.ReqSyncGame{RoundId: "-1"}); err != nil {
			logger.Panic("majSoul SyncGame error.", zap.Error(err))
		} else {
			logger.Debug("majSoul SyncGame.", zap.Reflect("resSyncGame", resSyncGame))
		}
	}

	{ // 获取当前玩家状态
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if _, err := majSoul.FastTestClient.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
			logger.Panic("majSoul FetchGamePlayerState error.", zap.Error(err))
		} else {
			logger.Debug("majSoul FetchGamePlayerState.")
		}
	}

	{ // 告诉服务器状态同步已经完成
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if _, err := majSoul.FastTestClient.FinishSyncGame(ctx, &message.ReqCommon{}); err != nil {
			logger.Panic("majSoul FinishSyncGame error.", zap.Error(err))
		} else {
			logger.Debug("majSoul FinishSyncGame.")
		}
	}

	{ // 获取当前玩家状态
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if _, err := majSoul.FastTestClient.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
			logger.Panic("majSoul FetchGamePlayerState error.", zap.Error(err))
		} else {
			logger.Debug("majSoul FetchGamePlayerState.")
		}
	}
}

func UpdateLoginInfo(majSoul *majsoul.MajSoul) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resFetchLastPrivacy, err := majSoul.LobbyClient.FetchLastPrivacy(ctx, &message.ReqFetchLastPrivacy{})
	if err != nil {
		logger.Panic("majSoul FetchLastPrivacy error.", zap.Error(err))
	}
	logger.Info("majSoul FetchLastPrivacy.", zap.Reflect("resFetchLastPrivacy", resFetchLastPrivacy))

	resFetchServerTime, err := majSoul.LobbyClient.FetchServerTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchServerTime error.", zap.Error(err))
	}
	logger.Info("majSoul FetchServerTime.", zap.Reflect("resFetchServerTime", resFetchServerTime))

	resServerSettings, err := majSoul.LobbyClient.FetchServerSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchServerSettings error.", zap.Error(err))
	}
	logger.Info("majSoul FetchServerSettings.", zap.Reflect("resServerSettings", resServerSettings))

	resConnectionInfo, err := majSoul.LobbyClient.FetchConnectionInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchConnectionInfo error.", zap.Error(err))
	}
	logger.Info("majSoul FetchConnectionInfo.", zap.Reflect("resConnectionInfo", resConnectionInfo))

	resClientValue, err := majSoul.LobbyClient.FetchClientValue(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchClientValue error.", zap.Error(err))
	}
	logger.Info("majSoul FetchClientValue.", zap.Reflect("resClientValue", resClientValue))

	resFriendList, err := majSoul.LobbyClient.FetchFriendList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchFriendList error.", zap.Error(err))
	}
	logger.Info("majSoul FetchFriendList.", zap.Reflect("resFriendList", resFriendList))

	resFriendApplyList, err := majSoul.LobbyClient.FetchFriendApplyList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchFriendApplyList error.", zap.Error(err))
	}
	logger.Info("majSoul FetchFriendApplyList.", zap.Reflect("resFriendApplyList", resFriendApplyList))

	resFetchrecentFriend, err := majSoul.LobbyClient.FetchRecentFriend(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchRecentFriend.", zap.Error(err))
	}
	logger.Info("majSoul FetchRecentFriend.", zap.Reflect("resFetchrecentFriend", resFetchrecentFriend))

	resMailInfo, err := majSoul.LobbyClient.FetchMailInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchMailInfo error.", zap.Error(err))
	}
	logger.Info("majSoul FetchMailInfo.", zap.Reflect("resMailInfo", resMailInfo))

	resDailyTask, err := majSoul.LobbyClient.FetchDailyTask(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchDailyTask error.", zap.Error(err))
	}
	logger.Info("majSoul FetchDailyTask.", zap.Reflect("resDailyTask", resDailyTask))

	resReviveCoinInfo, err := majSoul.LobbyClient.FetchReviveCoinInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchReviveCoinInfo error.", zap.Error(err))
	}
	logger.Info("majSoul FetchReviveCoinInfo.", zap.Reflect("resReviveCoinInfo", resReviveCoinInfo))

	resTitleList, err := majSoul.LobbyClient.FetchTitleList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchTitleList error.", zap.Error(err))
	}
	logger.Info("majSoul FetchTitleList.", zap.Reflect("resTitleList", resTitleList))

	resBagInfo, err := majSoul.LobbyClient.FetchBagInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchBagInfo error.", zap.Error(err))
	}
	logger.Info("majSoul FetchBagInfo.", zap.Reflect("resBagInfo", resBagInfo))

	resShopInfo, err := majSoul.LobbyClient.FetchShopInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchShopInfo error.", zap.Error(err))
	}
	logger.Info("majSoul FetchShopInfo.", zap.Reflect("resShopInfo", resShopInfo))

	resFetchShopInterval, err := majSoul.LobbyClient.FetchShopInterval(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchShopInterval error.", zap.Error(err))
	}
	logger.Info("majSoul FetchShopInterval.", zap.Reflect("resFetchShopInterval", resFetchShopInterval))

	resActivityList, err := majSoul.LobbyClient.FetchActivityList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchActivityList error.", zap.Error(err))
	}
	logger.Info("majSoul FetchActivityList.", zap.Reflect("resActivityList", resActivityList))

	resAccountActivityData, err := majSoul.LobbyClient.FetchAccountActivityData(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchAccountActivityData error.", zap.Error(err))
	}
	logger.Info("majSoul FetchAccountActivityData.", zap.Reflect("resAccountActivityData", resAccountActivityData))

	resFetchActivityInterval, err := majSoul.LobbyClient.FetchActivityInterval(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchActivityInterval error.", zap.Error(err))
	}
	logger.Info("majSoul FetchActivityInterval.", zap.Reflect("resFetchActivityInterval", resFetchActivityInterval))

	resActivityBuff, err := majSoul.LobbyClient.FetchActivityBuff(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchActivityBuff error.", zap.Error(err))
	}
	logger.Info("majSoul FetchActivityBuff.", zap.Reflect("resActivityBuff", resActivityBuff))

	resVipReward, err := majSoul.LobbyClient.FetchVipReward(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchVipReward error.", zap.Error(err))
	}
	logger.Info("majSoul FetchVipReward.", zap.Reflect("resVipReward", resVipReward))

	resMonthTicketInfo, err := majSoul.LobbyClient.FetchMonthTicketInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchMonthTicketInfo error.", zap.Error(err))
	}
	logger.Info("majSoul FetchMonthTicketInfo.", zap.Reflect("resMonthTicketInfo", resMonthTicketInfo))

	resAchievement, err := majSoul.LobbyClient.FetchAchievement(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchAchievement error.", zap.Error(err))
	}
	logger.Info("majSoul FetchAchievement.", zap.Reflect("resAchievement", resAchievement))

	resCommentSetting, err := majSoul.LobbyClient.FetchCommentSetting(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchCommentSetting error.", zap.Error(err))
	}
	logger.Info("majSoul FetchCommentSetting.", zap.Reflect("resCommentSetting", resCommentSetting))

	resAccountSettings, err := majSoul.LobbyClient.FetchAccountSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchAccountSettings error.", zap.Error(err))
	}
	logger.Info("majSoul FetchAccountSettings.", zap.Reflect("resAccountSettings", resAccountSettings))

	resModNicknameTime, err := majSoul.LobbyClient.FetchModNicknameTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchModNicknameTime error.", zap.Error(err))
	}
	logger.Info("majSoul FetchModNicknameTime.", zap.Reflect("resModNicknameTime", resModNicknameTime))

	resMisc, err := majSoul.LobbyClient.FetchMisc(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchMisc error.", zap.Error(err))
	}
	logger.Info("majSoul FetchMisc.", zap.Reflect("resMisc", resMisc))

	resAnnouncement, err := majSoul.LobbyClient.FetchAnnouncement(ctx, &message.ReqFetchAnnouncement{})
	if err != nil {
		logger.Panic("majSoul FetchAnnouncement error.", zap.Error(err))
	}
	logger.Info("majSoul FetchAnnouncement.", zap.Reflect("resAnnouncement", resAnnouncement))

	// 写错了吧 req?
	reqRollingNotice, err := majSoul.LobbyClient.FetchRollingNotice(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul FetchRollingNotice error.", zap.Error(err))
	}
	logger.Info("majSoul FetchRollingNotice.", zap.Reflect("reqRollingNotice", reqRollingNotice))

	resCommon, err := majSoul.LobbyClient.LoginSuccess(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Panic("majSoul LoginSuccess error.", zap.Error(err))
	}
	logger.Info("majSoul LoginSuccess.", zap.Reflect("resCommon", resCommon))
}

func main() {
	defer logger.Init()()

	account, exists := os.LookupEnv("account")
	if !exists {
		panic("account is required.")
	}

	password, exists := os.LookupEnv("password")
	if !exists {
		panic("password is required.")
	}

	majSoul := majsoul.NewMajSoul(&majsoul.Config{ProxyAddress: ""})
	{ // 寻找可用服务器
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := majSoul.LookupGateway(ctx, majsoul.ServerAddressList)
		if err != nil {
			panic(err)
		}
	}

	gameState := &GameState{}

	{ // 登录
		var resLogin *message.ResLogin
		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			var err error
			resLogin, err = majSoul.Login(ctx, account, password)
			if err != nil {
				panic(err)
			}
			if resLogin.Error != nil {
				panic(resLogin.Error)
			}
		}

		gameState.account = resLogin.Account
		gameState.accessToken = resLogin.AccessToken
		if resLogin.GameInfo != nil {
			gameState.connectToken = resLogin.GameInfo.ConnectToken
			gameState.gameUuid = resLogin.GameInfo.GameUuid
		}

		// 模拟雀魂游览器进行的请求
		UpdateLoginInfo(majSoul)

		// 尝试重连到游戏中
		ReConnGame(gameState, majSoul, resLogin)
	}

	// 添加网关重连时进行的操作
	majSoul.OnGatewayReconnect(func() {
		if len(gameState.accessToken) == 0 {
			panic(fmt.Sprintf(""))
		}
		var err error
		{ // 检查 token 是否可用
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			_, err = majSoul.LobbyClient.Oauth2Check(ctx, &message.ReqOauth2Check{AccessToken: gameState.accessToken})
			if err != nil {
				panic(fmt.Sprintf("gateway Oauth2Check error %v", err))
			}
		}
		{ //
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			resLogin, err := majSoul.LobbyClient.Oauth2Login(ctx, &message.ReqOauth2Login{
				AccessToken: gameState.accessToken,
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
			if resLogin.Error != nil {
				panic(resLogin.Error)
			}
			gameState.account = resLogin.Account
			gameState.accessToken = resLogin.AccessToken
			if resLogin.GameInfo != nil {
				gameState.connectToken = resLogin.GameInfo.ConnectToken
				gameState.gameUuid = resLogin.GameInfo.GameUuid
			}
			UpdateLoginInfo(majSoul)
			ReConnGame(gameState, majSoul, resLogin)
		}
	})

	// 添加游戏服务器重连时进行的操作
	majSoul.OnGameReconnect(func() {
		{ // 验证用户身份
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

		{ // 同步游戏状态
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

		{ // 同步玩家状态
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if _, err := majSoul.FastTestClient.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
				logger.Error("majSoul FetchGamePlayerState error.", zap.Error(err))
				return
			} else {
				logger.Debug("majSoul FetchGamePlayerState.")
			}
		}

		{ // 完成游戏数据同步
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if _, err := majSoul.FastTestClient.FinishSyncGame(ctx, &message.ReqCommon{}); err != nil {
				logger.Error("majSoul FinishSyncGame error.", zap.Error(err))
				return
			} else {
				logger.Debug("majSoul FinishSyncGame.")
			}
		}

		{ // 同步玩家状态
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

	// 响应以下消息
	majSoul.Handle(
		gameState.NotifyClientMessage,
		gameState.NotifyFriendViewChange,
		gameState.NotifyEndGameVote,
		gameState.NotifyRoomGameStart,
		//gameState.ActionMJStart,
		//gameState.ActionNewRound,
		//gameState.ActionDealTile,
		//gameState.ActionDiscardTile,
		//gameState.ActionChiPengGang,
		//gameState.ActionAnGangAddGang,
		//gameState.ActionHule,
		//gameState.ActionLiuJu,
		//gameState.ActionNoTile,
	)

	logger.Debug("Game Startup")
	select {}
}
