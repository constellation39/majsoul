package main

import (
	"context"
	"flag"
	"time"

	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
)

type Majsoul struct {
	*majsoul.Majsoul
}

var (
	account  = flag.String("account", "", "majsoul login when the account(email or mobile number).")
	password = flag.String("password", "", "majsoul login when the password.")
)

func (majsoul *Majsoul) NotifyClientMessage(ctx context.Context, notify *message.NotifyClientMessage) {
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
	invitationRoom := new(InvitationRoom)
	err := json.Unmarshal([]byte(notify.Content), invitationRoom)
	if err != nil {
		logger.Error("Unmarshal", zap.Error(err))
		return
	}
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

	// 按照雀魂web端的请求进行模拟
	timeOutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	resLogin, err := client.Login(timeOutCtx, *account, *password)
	if err != nil {
		logger.Error("majsoul login error.", zap.Error(err))
		return
	}
	logger.Info("majsoul login.", zap.Reflect("resLogin", resLogin))

	err = UpdateLoginInfo(ctx, client)
	if err != nil {
		return
	}
}

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
