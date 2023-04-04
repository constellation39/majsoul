package main

import (
	"context"
	"encoding/json"
	"flag"
	"time"

	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
)

var (
	account      = flag.String("account", "", "majsoul login when the account(email or mobile number).")
	password     = flag.String("password", "", "majsoul login when the password.")
	serverProxy  = flag.String("serverProxy", "", "majsoul request server when the proxy.")
	gatewayProxy = flag.String("gatewayProxy", "", "majsoul connect gateway when the proxy.")
	gameProxy    = flag.String("gameProxy", "", "majsoul connect game when the proxy.")
)

type Majsoul struct {
	*majsoul.Majsoul
	seat uint32
}

func NewMajSoul(ctx context.Context) (*Majsoul, error) {
	configOptions := make([]majsoul.ConfigOption, 0, 2)

	if len(*serverProxy) > 0 {
		configOptions = append(configOptions, majsoul.WithServerProxy(*serverProxy))
	}

	if len(*gatewayProxy) > 0 {
		configOptions = append(configOptions, majsoul.WithGatewayProxy(*gatewayProxy))
	}

	if len(*gameProxy) > 0 {
		configOptions = append(configOptions, majsoul.WithGameProxy(*gameProxy))
	}

	// 初始化一个客户端
	subClient := majsoul.New(configOptions...)

	if err := subClient.TryConnect(ctx, majsoul.ServerAddressList); err != nil {
		return nil, err
	}

	client := &Majsoul{Majsoul: subClient}
	// Majsoul 是一个处理麻将游戏逻辑的结构体。要使用它，请先创建一个 Majsoul 对象，
	// 需要监听雀魂服务器下发通知时，需要实现这个接口 majsoul.Implement
	// majsoul.Majsoul 原生实现了这个接口，只需要重写需要的方法即可
	client.Implement(client)
	logger.Info("majsoul client is created.", zap.Reflect("ServerAddress", subClient.ServerAddress))
	return client, nil
}

// NotifyClientMessage 客户端消息
// message.NotifyClientMessage filed Type == 1 时为受到邀请
// note: 这个函数的只实现了接受到邀请的通知
func (mSoul *Majsoul) NotifyClientMessage(ctx context.Context, notify *message.NotifyClientMessage) {
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
	if notify.Type != 1 {
		logger.Info("notify.Type != -1", zap.Uint32("type", notify.Type))
		return
	}
	invitationRoom := new(InvitationRoom)
	err := json.Unmarshal([]byte(notify.Content), invitationRoom)
	if err != nil {
		logger.Error("Unmarshal", zap.Error(err))
		return
	}

	// 加入房间
	_, err = mSoul.JoinRoom(ctx, &message.ReqJoinRoom{
		RoomId:              invitationRoom.RoomID,
		ClientVersionString: mSoul.Version.Web(),
	})
	if err != nil {
		logger.Error("JoinRoom", zap.Error(err))
		return
	}

	time.Sleep(time.Second)

	// 准备
	_, err = mSoul.ReadyPlay(ctx, &message.ReqRoomReady{Ready: true})
	if err != nil {
		logger.Error("ReadyPlay", zap.Error(err))
		return
	}
}

// NotifyEndGameVote 有人发起投降
func (mSoul *Majsoul) NotifyEndGameVote(ctx context.Context, notify *message.NotifyEndGameVote) {
	_, err := mSoul.VoteGameEnd(ctx, &message.ReqVoteGameEnd{Yes: true})
	if err != nil {
		logger.Error("VoteGameEnd", zap.Error(err))
	}
}

// 从等待房间进入游戏时
func (mSoul *Majsoul) NotifyRoomGameStart(ctx context.Context, notify *message.NotifyRoomGameStart) {
	// 记录自己的座位号
	for i, uid := range mSoul.GameInfo.SeatList {
		if uid == mSoul.Account.AccountId {
			mSoul.seat = uint32(i)
			break
		}
	}
}

// ActionMJStart 游戏开始
func (mSoul *Majsoul) ActionMJStart(context.Context, *message.ActionMJStart) {
}

// ActionNewRound 回合开始
func (mSoul *Majsoul) ActionNewRound(ctx context.Context, action *message.ActionNewRound) {
	// 如果是庄家
	if len(action.Tiles) != 14 {
		return
	}
	tile13 := action.Tiles[13]
	time.Sleep(time.Second * 3)
	_, err := mSoul.InputOperation(ctx, &message.ReqSelfOperation{
		Type:    majsoul.ActionDiscard,
		Tile:    tile13,
		Moqie:   true,
		Timeuse: 1,
	})
	if err != nil {
		logger.Error("InputOperation failed", zap.Error(err))
	}
}

// ActionDealTile 摸牌
func (mSoul *Majsoul) ActionDealTile(ctx context.Context, action *message.ActionDealTile) {
	// 如果不是自己摸牌
	if action.Seat != mSoul.seat {
		return
	}

	if len(action.Tile) == 0 {
		logger.Error("摸牌是空的")
		return
	}

	time.Sleep(time.Second * 3)
	_, err := mSoul.InputOperation(ctx, &message.ReqSelfOperation{
		Type:    majsoul.ActionDiscard,
		Tile:    action.Tile,
		Moqie:   true,
		Timeuse: 1,
	})
	if err != nil {
		logger.Error("InputOperation failed", zap.Error(err))
	}

}

// ActionDiscardTile 打牌
func (mSoul *Majsoul) ActionDiscardTile(ctx context.Context, action *message.ActionDiscardTile) {
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
					_, err := mSoul.InputOperation(ctx, &message.ReqSelfOperation{
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

// ActionChiPengGang 吃碰杠的通知
func (mSoul *Majsoul) ActionChiPengGang(ctx context.Context, action *message.ActionChiPengGang) {
	switch action.Type {
	case majsoul.NotifyChi:
	case majsoul.NotifyPon:
	case majsoul.NotifyKan:
	}
}

// ActionAnGangAddGang 暗杠和加杠的通知
func (mSoul *Majsoul) ActionAnGangAddGang(ctx context.Context, action *message.ActionAnGangAddGang) {
	switch action.Type {
	case majsoul.NotifyAnKan:
	case majsoul.NotifyKaKan:
	}
}

func (mSoul *Majsoul) ActionHule(ctx context.Context, action *message.ActionHule) {
}

func (mSoul *Majsoul) ActionLiuJu(ctx context.Context, action *message.ActionLiuJu) {
}

func (mSoul *Majsoul) ActionNoTile(ctx context.Context, action *message.ActionNoTile) {
}

// ActionBaBei(ctx context.Context, action *message.ActionBaBei)
// ActionNewCard(ctx context.Context,action *message.ActionNewCard)
// ActionSelectGap(ctx context.Context,action *message.ActionSelectGap)
// ActionChangeTile(ctx context.Context,action *message.ActionChangeTile)
// ActionRevealTile(ctx context.Context,action *message.ActionRevealTile)
// ActionUnveilTile(ctx context.Context,action *message.ActionUnveilTile)
// ActionLockTile(ctx context.Context,action *message.ActionLockTile)
// ActionGangResult(ctx context.Context,action *message.ActionGangResult)
// ActionGangResultEnd(ctx context.Context,action *message.ActionGangResultEnd)
// ActionHuleXueZhanMid(ctx context.Context,action *message.ActionHuleXueZhanMid)
// ActionHuleXueZhanEnd(ctx context.Context,action *message.ActionHuleXueZhanEnd)

// 按照雀魂web端的请求进行模拟
func UpdateLoginInfo(ctx context.Context, client *Majsoul) error {
	resFetchLastPrivacy, err := client.FetchLastPrivacy(ctx, &message.ReqFetchLastPrivacy{})
	if err != nil {
		logger.Error("client FetchLastPrivacy error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchLastPrivacy.", zap.Reflect("resFetchLastPrivacy", resFetchLastPrivacy))

	resFetchServerTime, err := client.FetchServerTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchServerTime error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchServerTime.", zap.Reflect("resFetchServerTime", resFetchServerTime))

	resServerSettings, err := client.FetchServerSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchServerSettings error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchServerSettings.", zap.Reflect("resServerSettings", resServerSettings))

	resConnectionInfo, err := client.FetchConnectionInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchConnectionInfo error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchConnectionInfo.", zap.Reflect("resConnectionInfo", resConnectionInfo))

	resClientValue, err := client.FetchClientValue(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchClientValue error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchClientValue.", zap.Reflect("resClientValue", resClientValue))

	resFriendList, err := client.FetchFriendList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchFriendList error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchFriendList.", zap.Reflect("resFriendList", resFriendList))

	resFriendApplyList, err := client.FetchFriendApplyList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchFriendApplyList error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchFriendApplyList.", zap.Reflect("resFriendApplyList", resFriendApplyList))

	resFetchrecentFriend, err := client.FetchRecentFriend(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchRecentFriend.", zap.Error(err))
		return err
	}
	logger.Info("client FetchRecentFriend.", zap.Reflect("resFetchrecentFriend", resFetchrecentFriend))

	resMailInfo, err := client.FetchMailInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchMailInfo error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchMailInfo.", zap.Reflect("resMailInfo", resMailInfo))

	resDailyTask, err := client.FetchDailyTask(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchDailyTask error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchDailyTask.", zap.Reflect("resDailyTask", resDailyTask))

	resReviveCoinInfo, err := client.FetchReviveCoinInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchReviveCoinInfo error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchReviveCoinInfo.", zap.Reflect("resReviveCoinInfo", resReviveCoinInfo))

	resTitleList, err := client.FetchTitleList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchTitleList error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchTitleList.", zap.Reflect("resTitleList", resTitleList))

	resBagInfo, err := client.FetchBagInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchBagInfo error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchBagInfo.", zap.Reflect("resBagInfo", resBagInfo))

	resShopInfo, err := client.FetchShopInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchShopInfo error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchShopInfo.", zap.Reflect("resShopInfo", resShopInfo))

	resFetchShopInterval, err := client.FetchShopInterval(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchShopInterval error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchShopInterval.", zap.Reflect("resFetchShopInterval", resFetchShopInterval))

	resActivityList, err := client.FetchActivityList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchActivityList error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchActivityList.", zap.Reflect("resActivityList", resActivityList))

	resAccountActivityData, err := client.FetchAccountActivityData(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchAccountActivityData error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchAccountActivityData.", zap.Reflect("resAccountActivityData", resAccountActivityData))

	resFetchActivityInterval, err := client.FetchActivityInterval(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchActivityInterval error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchActivityInterval.", zap.Reflect("resFetchActivityInterval", resFetchActivityInterval))

	resActivityBuff, err := client.FetchActivityBuff(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchActivityBuff error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchActivityBuff.", zap.Reflect("resActivityBuff", resActivityBuff))

	resVipReward, err := client.FetchVipReward(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchVipReward error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchVipReward.", zap.Reflect("resVipReward", resVipReward))

	resMonthTicketInfo, err := client.FetchMonthTicketInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchMonthTicketInfo error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchMonthTicketInfo.", zap.Reflect("resMonthTicketInfo", resMonthTicketInfo))

	resAchievement, err := client.FetchAchievement(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchAchievement error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchAchievement.", zap.Reflect("resAchievement", resAchievement))

	resCommentSetting, err := client.FetchCommentSetting(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchCommentSetting error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchCommentSetting.", zap.Reflect("resCommentSetting", resCommentSetting))

	resAccountSettings, err := client.FetchAccountSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchAccountSettings error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchAccountSettings.", zap.Reflect("resAccountSettings", resAccountSettings))

	resModNicknameTime, err := client.FetchModNicknameTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchModNicknameTime error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchModNicknameTime.", zap.Reflect("resModNicknameTime", resModNicknameTime))

	resMisc, err := client.FetchMisc(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchMisc error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchMisc.", zap.Reflect("resMisc", resMisc))

	resAnnouncement, err := client.FetchAnnouncement(ctx, &message.ReqFetchAnnouncement{})
	if err != nil {
		logger.Error("client FetchAnnouncement error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchAnnouncement.", zap.Reflect("resAnnouncement", resAnnouncement))

	// 写错了吧 req?
	reqRollingNotice, err := client.FetchRollingNotice(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client FetchRollingNotice error.", zap.Error(err))
		return err
	}
	logger.Info("client FetchRollingNotice.", zap.Reflect("reqRollingNotice", reqRollingNotice))

	resCommon, err := client.LoginSuccess(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("client LoginSuccess error.", zap.Error(err))
		return err
	}
	logger.Info("client LoginSuccess.", zap.Reflect("resCommon", resCommon))

	return nil
}

func main() {
	flag.Parse()
	logger.SetOutput("stdout")
	logger.SetErrorOutput("stderr")
	logger.EnableDevelopment()

	if *account == "" {
		logger.Error("account is required.")
		return
	}

	if *password == "" {
		logger.Error("password is required.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := NewMajSoul(ctx)
	if err != nil {
		logger.Error("client client is not created.", zap.Error(err))
		return
	}

	timeOutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	resLogin, err := client.Login(timeOutCtx, *account, *password)
	cancel()
	if err != nil {
		logger.Error("client Login error.", zap.Error(err))
		return
	}
	if resLogin.Error != nil && resLogin.Error.Code != 0 {
		errorString := majsoul.ErrorString(resLogin.Error)
		logger.Error("client Login error.", zap.Uint32("Code", resLogin.Error.Code), zap.String("errorString", errorString))
		return
	}
	logger.Info("client Login.", zap.Reflect("resLogin", resLogin))

	err = UpdateLoginInfo(ctx, client)
	if err != nil {
		logger.Error("UpdateLoginInfo error.", zap.Error(err))
		return
	}

	// 重连到正在进行对局的游戏中
	if resLogin.Account != nil && resLogin.Account.RoomId != 0 {
		if err := client.ConnGame(ctx); err != nil {
			logger.Error("client ConnGame error.", zap.Error(err))
		}

		var err error
		client.GameInfo, err = client.AuthGame(ctx, &message.ReqAuthGame{
			AccountId: client.Account.AccountId,
			Token:     resLogin.GameInfo.ConnectToken,
			GameUuid:  resLogin.GameInfo.GameUuid,
		})
		if err != nil {
			logger.Error("client AuthGame error.", zap.Error(err))
		}

		for i, uid := range client.GameInfo.SeatList {
			if uid == client.Account.AccountId {
				client.seat = uint32(i)
				break
			}
		}

		if resSyncGame, err := client.SyncGame(ctx, &message.ReqSyncGame{RoundId: "-1"}); err != nil {
			logger.Error("client SyncGame error.", zap.Error(err))
		} else {
			logger.Debug("client SyncGame.", zap.Reflect("resSyncGame", resSyncGame))
		}

		if _, err := client.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
			logger.Error("client FetchGamePlayerState error.", zap.Error(err))
		} else {
			logger.Debug("client FetchGamePlayerState.")
		}

		if _, err := client.FinishSyncGame(ctx, &message.ReqCommon{}); err != nil {
			logger.Error("client FinishSyncGame error.", zap.Error(err))
		} else {
			logger.Debug("client FinishSyncGame.")
		}

		if _, err := client.FetchGamePlayerState(ctx, &message.ReqCommon{}); err != nil {
			logger.Error("client FetchGamePlayerState error.", zap.Error(err))
		} else {
			logger.Debug("client FetchGamePlayerState.")
		}
	}

	client.OnGameReconnect(func(ctx context.Context, resSyncGame *message.ResSyncGame) {
		if client.GameInfo == nil {
			logger.Error("client.GameInfo is nil.")
			return
		}
		for i, uid := range client.GameInfo.SeatList {
			if uid == client.Account.AccountId {
				client.seat = uint32(i)
				break
			}
		}

		if resSyncGame.GameRestore == nil {
			return
		}

		if resSyncGame.GameRestore.Actions == nil {
			return
		}

		for _, action := range resSyncGame.GameRestore.Actions {
			actionMessage, err := majsoul.ActionFromActionPrototype(action)
			if err != nil {
				logger.Debug("client ActionFromActionPrototype error", zap.Error(err))
				continue
			}
			logger.Debug("client resSyncGame aciton", zap.Reflect("actionMessage", actionMessage))
		}

	})

	<-ctx.Done()
}
