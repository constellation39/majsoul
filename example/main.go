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

type Game struct {
	seat         uint32
	account      *message.Account     // 该字段应在登录成功后访问
	gameInfo     *message.ResAuthGame // 该字段应在进入游戏桌面后访问
	accessToken  string               // 验证身份时使用 的 token
	connectToken string               // 重连时使用的 token
	gameUuid     string               // 是否在游戏中
}

func (Game) NotifyClientMessage(majSoul *majsoul.MajSoul, notifyClientMessage *message.NotifyClientMessage) {
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

func (Game) NotifyFriendViewChange(majSoul *majsoul.MajSoul, notifyFriendViewChange *message.NotifyFriendViewChange) {
	logger.Debug("", zap.Reflect("notifyFriendViewChange", notifyFriendViewChange))
}

// NotifyEndGameVote 有人发起投降
func (Game) NotifyEndGameVote(majSoul *majsoul.MajSoul, notifyEndGameVote *message.NotifyEndGameVote) {
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
func (game *Game) NotifyRoomGameStart(majSoul *majsoul.MajSoul, notifyRoomGameStart *message.NotifyRoomGameStart) {

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
		game.gameInfo, err = majSoul.FastTestClient.AuthGame(ctx, &message.ReqAuthGame{
			AccountId: game.account.AccountId,
			Token:     game.connectToken,
			GameUuid:  game.gameUuid,
		})
		if err != nil {
			logger.Error("majsoul NotifyRoomGameStart AuthGame error: ", zap.Error(err))
			return
		}
	}
	game.connectToken = notifyRoomGameStart.ConnectToken
	game.gameUuid = notifyRoomGameStart.GameUuid
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
	for i, uid := range game.gameInfo.SeatList {
		if uid == game.account.AccountId {
			game.seat = uint32(i)
			break
		}
	}
}

// ActionMJStart 游戏开始
func (Game) ActionMJStart(majSoul *majsoul.MajSoul, actionMJStart *message.ActionMJStart) {
}

// ActionNewRound 回合开始
func (Game) ActionNewRound(majSoul *majsoul.MajSoul, action *message.ActionNewRound) {
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
func (game *Game) ActionDealTile(majSoul *majsoul.MajSoul, action *message.ActionDealTile) {
	// 如果不是自己摸牌
	if action.Seat != game.seat {
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
func (Game) ActionDiscardTile(majSoul *majsoul.MajSoul, action *message.ActionDiscardTile) {
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
func (Game) ActionChiPengGang(majSoul *majsoul.MajSoul, action *message.ActionChiPengGang) {
	switch action.Type {
	case majsoul.NotifyChi:
	case majsoul.NotifyPon:
	case majsoul.NotifyKan:
	}
}

// ActionAnGangAddGang 暗杠和加杠的通知
func (Game) ActionAnGangAddGang(majSoul *majsoul.MajSoul, action *message.ActionAnGangAddGang) {
	switch action.Type {
	case majsoul.NotifyAnKan:
	case majsoul.NotifyKaKan:
	}
}

func (Game) ActionHule(majSoul *majsoul.MajSoul, action *message.ActionHule) {
}

func (Game) ActionLiuJu(majSoul *majsoul.MajSoul, action *message.ActionLiuJu) {
}

func (Game) ActionNoTile(majSoul *majsoul.MajSoul, action *message.ActionNoTile) {
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

	var game Game

	majSoul.Handle(
		game.NotifyClientMessage,
		game.NotifyFriendViewChange,
		game.NotifyEndGameVote,
		game.NotifyRoomGameStart,
		game.ActionMJStart,
		game.ActionNewRound,
		game.ActionDealTile,
		game.ActionDiscardTile,
		game.ActionChiPengGang,
		game.ActionAnGangAddGang,
		game.ActionHule,
		game.ActionLiuJu,
		game.ActionNoTile,
	)

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		resLogin, err := majSoul.Login(ctx, account, password)
		if err != nil {
			panic(err)
		}
		if resLogin.Error == nil {
			game.account = resLogin.Account
			game.accessToken = resLogin.AccessToken
			if resLogin.GameInfo != nil {
				game.connectToken = resLogin.GameInfo.ConnectToken
				game.gameUuid = resLogin.GameInfo.GameUuid
			}
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
				game.gameInfo, err = majSoul.FastTestClient.AuthGame(ctx, &message.ReqAuthGame{
					AccountId: resLogin.Account.AccountId,
					Token:     resLogin.GameInfo.ConnectToken,
					GameUuid:  resLogin.GameInfo.GameUuid,
				})
				if err != nil {
					logger.Error("client AuthGame error.", zap.Error(err))
				}
			}

			for i, uid := range game.gameInfo.SeatList {
				if uid == resLogin.Account.AccountId {
					game.seat = uint32(i)
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
		if len(game.accessToken) == 0 {
			panic(fmt.Sprintf(""))
		}
		var err error
		{
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			_, err = majSoul.LobbyClient.Oauth2Check(ctx, &message.ReqOauth2Check{AccessToken: game.accessToken})
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
				AccountId: game.account.AccountId,
				Token:     game.connectToken,
				GameUuid:  game.gameUuid,
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
