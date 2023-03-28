package majsoul

import (
	"context"

	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// IFNotify is the interface that must be implemented by a receiver.

func (majsoul *Majsoul) NotifyCaptcha(ctx context.Context, notify *message.NotifyCaptcha) {
	logger.Debug("NotifyCaptcha", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyRoomGameStart(ctx context.Context, notify *message.NotifyRoomGameStart) {
	logger.Debug("NotifyRoomGameStart", zap.Reflect("notify", notify))
	majsoul.ConnGame(ctx)
	var err error
	majsoul.GameInfo, err = majsoul.AuthGame(ctx, &message.ReqAuthGame{
		AccountId: majsoul.Account.AccountId,
		Token:     notify.ConnectToken,
		GameUuid:  notify.GameUuid,
	})
	if err != nil {
		logger.Error("NotifyRoomGameStart AuthGame error: ", zap.Error(err))
		return
	}

	_, err = majsoul.EnterGame(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("NotifyRoomGameStart EnterGame error:", zap.Error(err))
		return
	}
}

func (majsoul *Majsoul) NotifyMatchGameStart(ctx context.Context, notify *message.NotifyMatchGameStart) {
	logger.Debug("NotifyMatchGameStart", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyRoomPlayerReady(ctx context.Context, notify *message.NotifyRoomPlayerReady) {
	logger.Debug("NotifyRoomPlayerReady", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyRoomPlayerDressing(ctx context.Context, notify *message.NotifyRoomPlayerDressing) {
	logger.Debug("NotifyRoomPlayerDressing", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyRoomPlayerUpdate(ctx context.Context, notify *message.NotifyRoomPlayerUpdate) {
	logger.Debug("NotifyRoomPlayerUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyRoomKickOut(ctx context.Context, notify *message.NotifyRoomKickOut) {
	logger.Debug("NotifyRoomKickOut", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyFriendStateChange(ctx context.Context, notify *message.NotifyFriendStateChange) {
	logger.Debug("NotifyFriendStateChange", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyFriendViewChange(ctx context.Context, notify *message.NotifyFriendViewChange) {
	logger.Debug("NotifyFriendViewChange", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyFriendChange(ctx context.Context, notify *message.NotifyFriendChange) {
	logger.Debug("NotifyFriendChange", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyNewFriendApply(ctx context.Context, notify *message.NotifyNewFriendApply) {
	logger.Debug("NotifyNewFriendApply", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyClientMessage(ctx context.Context, notify *message.NotifyClientMessage) {
	logger.Debug("NotifyClientMessage", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAccountUpdate(ctx context.Context, notify *message.NotifyAccountUpdate) {
	logger.Debug("NotifyAccountUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAnotherLogin(ctx context.Context, notify *message.NotifyAnotherLogin) {
	logger.Debug("NotifyAnotherLogin", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAccountLogout(ctx context.Context, notify *message.NotifyAccountLogout) {
	logger.Debug("NotifyAccountLogout", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAnnouncementUpdate(ctx context.Context, notify *message.NotifyAnnouncementUpdate) {
	logger.Debug("NotifyAnnouncementUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyNewMail(ctx context.Context, notify *message.NotifyNewMail) {
	logger.Debug("NotifyNewMail", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyDeleteMail(ctx context.Context, notify *message.NotifyDeleteMail) {
	logger.Debug("NotifyDeleteMail", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyReviveCoinUpdate(ctx context.Context, notify *message.NotifyReviveCoinUpdate) {
	logger.Debug("NotifyReviveCoinUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyDailyTaskUpdate(ctx context.Context, notify *message.NotifyDailyTaskUpdate) {
	logger.Debug("NotifyDailyTaskUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityTaskUpdate(ctx context.Context, notify *message.NotifyActivityTaskUpdate) {
	logger.Debug("NotifyActivityTaskUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityPeriodTaskUpdate(ctx context.Context, notify *message.NotifyActivityPeriodTaskUpdate) {
	logger.Debug("NotifyActivityPeriodTaskUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAccountRandomTaskUpdate(ctx context.Context, notify *message.NotifyAccountRandomTaskUpdate) {
	logger.Debug("NotifyAccountRandomTaskUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivitySegmentTaskUpdate(ctx context.Context, notify *message.NotifyActivitySegmentTaskUpdate) {
	logger.Debug("NotifyActivitySegmentTaskUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityUpdate(ctx context.Context, notify *message.NotifyActivityUpdate) {
	logger.Debug("NotifyActivityUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAccountChallengeTaskUpdate(ctx context.Context, notify *message.NotifyAccountChallengeTaskUpdate) {
	logger.Debug("NotifyAccountChallengeTaskUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyNewComment(ctx context.Context, notify *message.NotifyNewComment) {
	logger.Debug("NotifyNewComment", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyRollingNotice(ctx context.Context, notify *message.NotifyRollingNotice) {
	logger.Debug("NotifyRollingNotice", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGiftSendRefresh(ctx context.Context, notify *message.NotifyGiftSendRefresh) {
	logger.Debug("NotifyGiftSendRefresh", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyShopUpdate(ctx context.Context, notify *message.NotifyShopUpdate) {
	logger.Debug("NotifyShopUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyVipLevelChange(ctx context.Context, notify *message.NotifyVipLevelChange) {
	logger.Debug("NotifyVipLevelChange", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyServerSetting(ctx context.Context, notify *message.NotifyServerSetting) {
	logger.Debug("NotifyServerSetting", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyPayResult(ctx context.Context, notify *message.NotifyPayResult) {
	logger.Debug("NotifyPayResult", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyCustomContestAccountMsg(ctx context.Context, notify *message.NotifyCustomContestAccountMsg) {
	logger.Debug("NotifyCustomContestAccountMsg", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyCustomContestSystemMsg(ctx context.Context, notify *message.NotifyCustomContestSystemMsg) {
	logger.Debug("NotifyCustomContestSystemMsg", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyMatchTimeout(ctx context.Context, notify *message.NotifyMatchTimeout) {
	logger.Debug("NotifyMatchTimeout", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyCustomContestState(ctx context.Context, notify *message.NotifyCustomContestState) {
	logger.Debug("NotifyCustomContestState", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityChange(ctx context.Context, notify *message.NotifyActivityChange) {
	logger.Debug("NotifyActivityChange", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAFKResult(ctx context.Context, notify *message.NotifyAFKResult) {
	logger.Debug("NotifyAFKResult", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2(ctx context.Context, notify *message.NotifyGameFinishRewardV2) {
	logger.Debug("NotifyGameFinishRewardV2", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityRewardV2(ctx context.Context, notify *message.NotifyActivityRewardV2) {
	logger.Debug("NotifyActivityRewardV2", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityPointV2(ctx context.Context, notify *message.NotifyActivityPointV2) {
	logger.Debug("NotifyActivityPointV2", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyLeaderboardPointV2(ctx context.Context, notify *message.NotifyLeaderboardPointV2) {
	logger.Debug("NotifyLeaderboardPointV2", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyNewGame(ctx context.Context, notify *message.NotifyNewGame) {
	logger.Debug("NotifyNewGame", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyPlayerLoadGameReady(ctx context.Context, notify *message.NotifyPlayerLoadGameReady) {
	logger.Debug("NotifyPlayerLoadGameReady", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGameBroadcast(ctx context.Context, notify *message.NotifyGameBroadcast) {
	logger.Debug("NotifyGameBroadcast", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGameEndResult(ctx context.Context, notify *message.NotifyGameEndResult) {
	logger.Debug("NotifyGameEndResult", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGameTerminate(ctx context.Context, notify *message.NotifyGameTerminate) {
	logger.Debug("NotifyGameTerminate", zap.Reflect("notify", notify))
	majsoul.FastTestConn = nil
	majsoul.FastTestClient = nil
}

func (majsoul *Majsoul) NotifyPlayerConnectionState(ctx context.Context, notify *message.NotifyPlayerConnectionState) {
	logger.Debug("NotifyPlayerConnectionState", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAccountLevelChange(ctx context.Context, notify *message.NotifyAccountLevelChange) {
	logger.Debug("NotifyAccountLevelChange", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGameFinishReward(ctx context.Context, notify *message.NotifyGameFinishReward) {
	logger.Debug("NotifyGameFinishReward", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityReward(ctx context.Context, notify *message.NotifyActivityReward) {
	logger.Debug("NotifyActivityReward", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityPoint(ctx context.Context, notify *message.NotifyActivityPoint) {
	logger.Debug("NotifyActivityPoint", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyLeaderboardPoint(ctx context.Context, notify *message.NotifyLeaderboardPoint) {
	logger.Debug("NotifyLeaderboardPoint", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGamePause(ctx context.Context, notify *message.NotifyGamePause) {
	logger.Debug("NotifyGamePause", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyEndGameVote(ctx context.Context, notify *message.NotifyEndGameVote) {
	logger.Debug("NotifyEndGameVote", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyObserveData(ctx context.Context, notify *message.NotifyObserveData) {
	logger.Debug("NotifyObserveData", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyRoomPlayerReady_AccountReadyState(ctx context.Context, notify *message.NotifyRoomPlayerReady_AccountReadyState) {
	logger.Debug("NotifyRoomPlayerReady_AccountReadyState", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyRoomPlayerDressing_AccountDressingState(ctx context.Context, notify *message.NotifyRoomPlayerDressing_AccountDressingState) {
	logger.Debug("NotifyRoomPlayerDressing_AccountDressingState", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyAnnouncementUpdate_AnnouncementUpdate(ctx context.Context, notify *message.NotifyAnnouncementUpdate_AnnouncementUpdate) {
	logger.Debug("NotifyAnnouncementUpdate_AnnouncementUpdate", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData(ctx context.Context, notify *message.NotifyActivityUpdate_FeedActivityData) {
	logger.Debug("NotifyActivityUpdate_FeedActivityData", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData_CountWithTimeData(ctx context.Context, notify *message.NotifyActivityUpdate_FeedActivityData_CountWithTimeData) {
	logger.Debug("NotifyActivityUpdate_FeedActivityData_CountWithTimeData", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData_GiftBoxData(ctx context.Context, notify *message.NotifyActivityUpdate_FeedActivityData_GiftBoxData) {
	logger.Debug("NotifyActivityUpdate_FeedActivityData_GiftBoxData", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyPayResult_ResourceModify(ctx context.Context, notify *message.NotifyPayResult_ResourceModify) {
	logger.Debug("NotifyPayResult_ResourceModify", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_LevelChange(ctx context.Context, notify *message.NotifyGameFinishRewardV2_LevelChange) {
	logger.Debug("NotifyGameFinishRewardV2_LevelChange", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_MatchChest(ctx context.Context, notify *message.NotifyGameFinishRewardV2_MatchChest) {
	logger.Debug("NotifyGameFinishRewardV2_MatchChest", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_MainCharacter(ctx context.Context, notify *message.NotifyGameFinishRewardV2_MainCharacter) {
	logger.Debug("NotifyGameFinishRewardV2_MainCharacter", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_CharacterGift(ctx context.Context, notify *message.NotifyGameFinishRewardV2_CharacterGift) {
	logger.Debug("NotifyGameFinishRewardV2_CharacterGift", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityRewardV2_ActivityReward(ctx context.Context, notify *message.NotifyActivityRewardV2_ActivityReward) {
	logger.Debug("NotifyActivityRewardV2_ActivityReward", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyActivityPointV2_ActivityPoint(ctx context.Context, notify *message.NotifyActivityPointV2_ActivityPoint) {
	logger.Debug("NotifyActivityPointV2_ActivityPoint", zap.Reflect("notify", notify))
}

func (majsoul *Majsoul) NotifyLeaderboardPointV2_LeaderboardPoint(ctx context.Context, notify *message.NotifyLeaderboardPointV2_LeaderboardPoint) {
	logger.Debug("NotifyLeaderboardPointV2_LeaderboardPoint", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyGameFinishReward_LevelChange(ctx context.Context, notify *message.NotifyGameFinishReward_LevelChange) {
	logger.Debug("NotifyGameFinishReward_LevelChange", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyGameFinishReward_MatchChest(ctx context.Context, notify *message.NotifyGameFinishReward_MatchChest) {
	logger.Debug("NotifyGameFinishReward_MatchChest", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyGameFinishReward_MainCharacter(ctx context.Context, notify *message.NotifyGameFinishReward_MainCharacter) {
	logger.Debug("NotifyGameFinishReward_MainCharacter", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyGameFinishReward_CharacterGift(ctx context.Context, notify *message.NotifyGameFinishReward_CharacterGift) {
	logger.Debug("NotifyGameFinishReward_CharacterGift", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyActivityReward_ActivityReward(ctx context.Context, notify *message.NotifyActivityReward_ActivityReward) {
	logger.Debug("NotifyActivityReward_ActivityReward", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyActivityPoint_ActivityPoint(ctx context.Context, notify *message.NotifyActivityPoint_ActivityPoint) {
	logger.Debug("NotifyActivityPoint_ActivityPoint", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyLeaderboardPoint_LeaderboardPoint(ctx context.Context, notify *message.NotifyLeaderboardPoint_LeaderboardPoint) {
	logger.Debug("NotifyLeaderboardPoint_LeaderboardPoint", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) NotifyEndGameVote_VoteResult(ctx context.Context, notify *message.NotifyEndGameVote_VoteResult) {
	logger.Debug("NotifyEndGameVote_VoteResult", zap.Reflect("notify", notify))

}

func (majsoul *Majsoul) PlayerLeaving(ctx context.Context, notify *message.PlayerLeaving) {
	logger.Debug("PlayerLeaving", zap.Reflect("notify", notify))
}

var keys = []int{0x84, 0x5e, 0x4e, 0x42, 0x39, 0xa2, 0x1f, 0x60, 0x1c}

func decode(data []byte) []byte {
	temp := make([]byte, len(data))
	copy(temp, data)
	for i := 0; i < len(temp); i++ {
		u := (23 ^ len(temp)) + 5*i + keys[i%len(keys)]&255
		temp[i] ^= byte(u)
	}
	return temp
}

func (majsoul *Majsoul) ActionPrototype(ctx context.Context, notify *message.ActionPrototype) {
	// logger.Debug("ActionPrototype", zap.Reflect("notify", notify))
	actionMessage := message.GetActionType(notify.Name)
	deData := decode(notify.Data)
	err := proto.Unmarshal(deData, actionMessage)
	if err != nil {
		logger.Error("ActionPrototype Unmarshal notify data failed: ", zap.Error(err), zap.Reflect("notify", notify), zap.Binary("data", notify.Data), zap.Binary("deCode", deData))
		return
	}
	switch notify.Name {
	case "ActionMJStart":
		majsoul.implement.ActionMJStart(ctx, actionMessage.(*message.ActionMJStart))
	case "ActionNewCard":
		majsoul.implement.ActionNewCard(ctx, actionMessage.(*message.ActionNewCard))
	case "ActionNewRound":
		majsoul.implement.ActionNewRound(ctx, actionMessage.(*message.ActionNewRound))
	case "ActionSelectGap":
		majsoul.implement.ActionSelectGap(ctx, actionMessage.(*message.ActionSelectGap))
	case "ActionChangeTile":
		majsoul.implement.ActionChangeTile(ctx, actionMessage.(*message.ActionChangeTile))
	case "ActionRevealTile":
		majsoul.implement.ActionRevealTile(ctx, actionMessage.(*message.ActionRevealTile))
	case "ActionUnveilTile":
		majsoul.implement.ActionUnveilTile(ctx, actionMessage.(*message.ActionUnveilTile))
	case "ActionLockTile":
		majsoul.implement.ActionLockTile(ctx, actionMessage.(*message.ActionLockTile))
	case "ActionDiscardTile":
		majsoul.implement.ActionDiscardTile(ctx, actionMessage.(*message.ActionDiscardTile))
	case "ActionDealTile":
		majsoul.implement.ActionDealTile(ctx, actionMessage.(*message.ActionDealTile))
	case "ActionChiPengGang":
		majsoul.implement.ActionChiPengGang(ctx, actionMessage.(*message.ActionChiPengGang))
	case "ActionGangResult":
		majsoul.implement.ActionGangResult(ctx, actionMessage.(*message.ActionGangResult))
	case "ActionGangResultEnd":
		majsoul.implement.ActionGangResultEnd(ctx, actionMessage.(*message.ActionGangResultEnd))
	case "ActionAnGangAddGang":
		majsoul.implement.ActionAnGangAddGang(ctx, actionMessage.(*message.ActionAnGangAddGang))
	case "ActionBaBei":
		majsoul.implement.ActionBaBei(ctx, actionMessage.(*message.ActionBaBei))
	case "ActionHule":
		majsoul.implement.ActionHule(ctx, actionMessage.(*message.ActionHule))
	case "ActionHuleXueZhanMid":
		majsoul.implement.ActionHuleXueZhanMid(ctx, actionMessage.(*message.ActionHuleXueZhanMid))
	case "ActionHuleXueZhanEnd":
		majsoul.implement.ActionHuleXueZhanEnd(ctx, actionMessage.(*message.ActionHuleXueZhanEnd))
	case "ActionLiuJu":
		majsoul.implement.ActionLiuJu(ctx, actionMessage.(*message.ActionLiuJu))
	case "ActionNoTile":
		majsoul.implement.ActionNoTile(ctx, actionMessage.(*message.ActionNoTile))
	default:
		logger.Error("unknown notify name: ", zap.String("name", notify.Name))
	}
}
