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

type Majsoul struct {
	*majsoul.Majsoul
	seat uint32
}

var (
	account  = flag.String("account", "", "majsoul login when the account(email or mobile number).")
	password = flag.String("password", "", "majsoul login when the password.")
)

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
		Fandian      int  `json:"fandian"`
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
		logger.Panic("InputOperation failed", zap.Error(err))
	}
}

// ActionDealTile 摸牌
func (mSoul *Majsoul) ActionDealTile(ctx context.Context, action *message.ActionDealTile) {
	// 如果不是自己摸牌
	if action.Seat != mSoul.seat {
		return
	}

	if action.Operation != nil && len(action.Operation.OperationList) != 0 {
		for _, operation := range action.Operation.OperationList {
			switch operation.Type {
			case majsoul.ActionDiscard:
				time.Sleep(time.Second * 3)
				_, err := mSoul.InputOperation(ctx, &message.ReqSelfOperation{
					Type:    majsoul.ActionDiscard,
					Tile:    action.Tile,
					Moqie:   true,
					Timeuse: 1,
				})
				if err != nil {
					logger.Panic("InputOperation failed", zap.Error(err))
				}
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
			}
		}
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
						logger.Panic("InputOperation failed", zap.Error(err))
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
		logger.Error("majsoul FetchLastPrivacy error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchLastPrivacy.", zap.Reflect("resFetchLastPrivacy", resFetchLastPrivacy))

	resFetchServerTime, err := client.FetchServerTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchServerTime error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchServerTime.", zap.Reflect("resFetchServerTime", resFetchServerTime))

	resServerSettings, err := client.FetchServerSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchServerSettings error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchServerSettings.", zap.Reflect("resServerSettings", resServerSettings))

	resConnectionInfo, err := client.FetchConnectionInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchConnectionInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchConnectionInfo.", zap.Reflect("resConnectionInfo", resConnectionInfo))

	resClientValue, err := client.FetchClientValue(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchClientValue error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchClientValue.", zap.Reflect("resClientValue", resClientValue))

	resFriendList, err := client.FetchFriendList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchFriendList error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchFriendList.", zap.Reflect("resFriendList", resFriendList))

	resFriendApplyList, err := client.FetchFriendApplyList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchFriendApplyList error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchFriendApplyList.", zap.Reflect("resFriendApplyList", resFriendApplyList))

	resFetchrecentFriend, err := client.FetchRecentFriend(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchRecentFriend.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchRecentFriend.", zap.Reflect("resFetchrecentFriend", resFetchrecentFriend))

	resMailInfo, err := client.FetchMailInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchMailInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchMailInfo.", zap.Reflect("resMailInfo", resMailInfo))

	resDailyTask, err := client.FetchDailyTask(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchDailyTask error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchDailyTask.", zap.Reflect("resDailyTask", resDailyTask))

	resReviveCoinInfo, err := client.FetchReviveCoinInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchReviveCoinInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchReviveCoinInfo.", zap.Reflect("resReviveCoinInfo", resReviveCoinInfo))

	resTitleList, err := client.FetchTitleList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchTitleList error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchTitleList.", zap.Reflect("resTitleList", resTitleList))

	resBagInfo, err := client.FetchBagInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchBagInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchBagInfo.", zap.Reflect("resBagInfo", resBagInfo))

	resShopInfo, err := client.FetchShopInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchShopInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchShopInfo.", zap.Reflect("resShopInfo", resShopInfo))

	resFetchShopInterval, err := client.FetchShopInterval(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchShopInterval error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchShopInterval.", zap.Reflect("resFetchShopInterval", resFetchShopInterval))

	resActivityList, err := client.FetchActivityList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchActivityList error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchActivityList.", zap.Reflect("resActivityList", resActivityList))

	resAccountActivityData, err := client.FetchAccountActivityData(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchAccountActivityData error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchAccountActivityData.", zap.Reflect("resAccountActivityData", resAccountActivityData))

	resFetchActivityInterval, err := client.FetchActivityInterval(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchActivityInterval error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchActivityInterval.", zap.Reflect("resFetchActivityInterval", resFetchActivityInterval))

	resActivityBuff, err := client.FetchActivityBuff(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchActivityBuff error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchActivityBuff.", zap.Reflect("resActivityBuff", resActivityBuff))

	resVipReward, err := client.FetchVipReward(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchVipReward error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchVipReward.", zap.Reflect("resVipReward", resVipReward))

	resMonthTicketInfo, err := client.FetchMonthTicketInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchMonthTicketInfo error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchMonthTicketInfo.", zap.Reflect("resMonthTicketInfo", resMonthTicketInfo))

	resAchievement, err := client.FetchAchievement(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchAchievement error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchAchievement.", zap.Reflect("resAchievement", resAchievement))

	resCommentSetting, err := client.FetchCommentSetting(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchCommentSetting error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchCommentSetting.", zap.Reflect("resCommentSetting", resCommentSetting))

	resAccountSettings, err := client.FetchAccountSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchAccountSettings error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchAccountSettings.", zap.Reflect("resAccountSettings", resAccountSettings))

	resModNicknameTime, err := client.FetchModNicknameTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchModNicknameTime error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchModNicknameTime.", zap.Reflect("resModNicknameTime", resModNicknameTime))

	resMisc, err := client.FetchMisc(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchMisc error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchMisc.", zap.Reflect("resMisc", resMisc))

	resAnnouncement, err := client.FetchAnnouncement(ctx, &message.ReqFetchAnnouncement{})
	if err != nil {
		logger.Error("majsoul FetchAnnouncement error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchAnnouncement.", zap.Reflect("resAnnouncement", resAnnouncement))

	// 写错了吧 req?
	reqRollingNotice, err := client.FetchRollingNotice(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchRollingNotice error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul FetchRollingNotice.", zap.Reflect("reqRollingNotice", reqRollingNotice))

	resCommon, err := client.LoginSuccess(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul LoginSuccess error.", zap.Error(err))
		return err
	}
	logger.Info("majsoul LoginSuccess.", zap.Reflect("resCommon", resCommon))

	return nil
}

func main() {
	flag.Parse()
	logger.EnableDevelopment()

	if *account == "" {
		logger.Error("account is required.")
		return
	}

	if *password == "" {
		logger.Error("password is required.")
		return
	}

	// 初始化一个客户端
	ctx := context.Background()
	subClient, err := majsoul.New(ctx)
	if err != nil {
		logger.Error("majsoul client is not created.", zap.Error(err))
		return
	}
	client := &Majsoul{Majsoul: subClient}
	// 使用了多态的方式实现
	// 需要监听雀魂服务器下发通知时，需要实现这个接口 majsoul.Implement
	// majsoul.Majsoul 原生实现了这个接口，只需要重写需要的方法即可
	subClient.Implement = client
	logger.Info("majsoul client is created.", zap.Reflect("ServerAddress", subClient.ServerAddress))

	timeOutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	resLogin, err := client.Login(timeOutCtx, *account, *password)
	if err != nil {
		logger.Error("majsoul login error.", zap.Error(err))
		return
	}
	if resLogin.Error != nil && resLogin.Error.Code != 0 {
		logger.Error("majsoul login error.", zap.Uint32("Code", resLogin.Error.Code))
	}
	logger.Info("majsoul login.", zap.Reflect("resLogin", resLogin))

	err = UpdateLoginInfo(ctx, client)
	if err != nil {
		return
	}

	// 检查是否在游戏中
	if resLogin.Account != nil && resLogin.Account.RoomId != 0 {
		err := client.ConnGame(ctx, resLogin.GameInfo.ConnectToken, resLogin.GameInfo.GameUuid)
		if err != nil {
			logger.Error("NotifyRoomGameStart ConnGame error: ", zap.Error(err))
			return
		}
		client.SyncGame(ctx, &message.ReqSyncGame{})
	}

	<-ctx.Done()
}
