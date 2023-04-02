package majsoul

import (
	"context"

	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Notify
// 雀魂proto协议中缺少描述监听消息的接口
// 故添加该接口，可能会丢失某些api
// 有没有更聪明点的办法？
type Notify interface {
	NotifyCaptcha(context.Context, *message.NotifyCaptcha)
	NotifyRoomGameStart(context.Context, *message.NotifyRoomGameStart)
	NotifyMatchGameStart(context.Context, *message.NotifyMatchGameStart)
	NotifyRoomPlayerReady(context.Context, *message.NotifyRoomPlayerReady)
	NotifyRoomPlayerDressing(context.Context, *message.NotifyRoomPlayerDressing)
	NotifyRoomPlayerUpdate(context.Context, *message.NotifyRoomPlayerUpdate)
	NotifyRoomKickOut(context.Context, *message.NotifyRoomKickOut)
	NotifyFriendStateChange(context.Context, *message.NotifyFriendStateChange)
	NotifyFriendViewChange(context.Context, *message.NotifyFriendViewChange)
	NotifyFriendChange(context.Context, *message.NotifyFriendChange)
	NotifyNewFriendApply(context.Context, *message.NotifyNewFriendApply)
	NotifyClientMessage(context.Context, *message.NotifyClientMessage)
	NotifyAccountUpdate(context.Context, *message.NotifyAccountUpdate)
	NotifyAnotherLogin(context.Context, *message.NotifyAnotherLogin)
	NotifyAccountLogout(context.Context, *message.NotifyAccountLogout)
	NotifyAnnouncementUpdate(context.Context, *message.NotifyAnnouncementUpdate)
	NotifyNewMail(context.Context, *message.NotifyNewMail)
	NotifyDeleteMail(context.Context, *message.NotifyDeleteMail)
	NotifyReviveCoinUpdate(context.Context, *message.NotifyReviveCoinUpdate)
	NotifyDailyTaskUpdate(context.Context, *message.NotifyDailyTaskUpdate)
	NotifyActivityTaskUpdate(context.Context, *message.NotifyActivityTaskUpdate)
	NotifyActivityPeriodTaskUpdate(context.Context, *message.NotifyActivityPeriodTaskUpdate)
	NotifyAccountRandomTaskUpdate(context.Context, *message.NotifyAccountRandomTaskUpdate)
	NotifyActivitySegmentTaskUpdate(context.Context, *message.NotifyActivitySegmentTaskUpdate)
	NotifyActivityUpdate(context.Context, *message.NotifyActivityUpdate)
	NotifyAccountChallengeTaskUpdate(context.Context, *message.NotifyAccountChallengeTaskUpdate)
	NotifyNewComment(context.Context, *message.NotifyNewComment)
	NotifyRollingNotice(context.Context, *message.NotifyRollingNotice)
	NotifyGiftSendRefresh(context.Context, *message.NotifyGiftSendRefresh)
	NotifyShopUpdate(context.Context, *message.NotifyShopUpdate)
	NotifyVipLevelChange(context.Context, *message.NotifyVipLevelChange)
	NotifyServerSetting(context.Context, *message.NotifyServerSetting)
	NotifyPayResult(context.Context, *message.NotifyPayResult)
	NotifyCustomContestAccountMsg(context.Context, *message.NotifyCustomContestAccountMsg)
	NotifyCustomContestSystemMsg(context.Context, *message.NotifyCustomContestSystemMsg)
	NotifyMatchTimeout(context.Context, *message.NotifyMatchTimeout)
	NotifyCustomContestState(context.Context, *message.NotifyCustomContestState)
	NotifyActivityChange(context.Context, *message.NotifyActivityChange)
	NotifyAFKResult(context.Context, *message.NotifyAFKResult)
	NotifyGameFinishRewardV2(context.Context, *message.NotifyGameFinishRewardV2)
	NotifyActivityRewardV2(context.Context, *message.NotifyActivityRewardV2)
	NotifyActivityPointV2(context.Context, *message.NotifyActivityPointV2)
	NotifyLeaderboardPointV2(context.Context, *message.NotifyLeaderboardPointV2)
	NotifyNewGame(context.Context, *message.NotifyNewGame)
	NotifyPlayerLoadGameReady(context.Context, *message.NotifyPlayerLoadGameReady)
	NotifyGameBroadcast(context.Context, *message.NotifyGameBroadcast)
	NotifyGameEndResult(context.Context, *message.NotifyGameEndResult)
	NotifyGameTerminate(context.Context, *message.NotifyGameTerminate)
	NotifyPlayerConnectionState(context.Context, *message.NotifyPlayerConnectionState)
	NotifyAccountLevelChange(context.Context, *message.NotifyAccountLevelChange)
	NotifyGameFinishReward(context.Context, *message.NotifyGameFinishReward)
	NotifyActivityReward(context.Context, *message.NotifyActivityReward)
	NotifyActivityPoint(context.Context, *message.NotifyActivityPoint)
	NotifyLeaderboardPoint(context.Context, *message.NotifyLeaderboardPoint)
	NotifyGamePause(context.Context, *message.NotifyGamePause)
	NotifyEndGameVote(context.Context, *message.NotifyEndGameVote)
	NotifyObserveData(context.Context, *message.NotifyObserveData)
	NotifyRoomPlayerReady_AccountReadyState(context.Context, *message.NotifyRoomPlayerReady_AccountReadyState)
	NotifyRoomPlayerDressing_AccountDressingState(context.Context, *message.NotifyRoomPlayerDressing_AccountDressingState)
	NotifyAnnouncementUpdate_AnnouncementUpdate(context.Context, *message.NotifyAnnouncementUpdate_AnnouncementUpdate)
	NotifyActivityUpdate_FeedActivityData(context.Context, *message.NotifyActivityUpdate_FeedActivityData)
	NotifyActivityUpdate_FeedActivityData_CountWithTimeData(context.Context, *message.NotifyActivityUpdate_FeedActivityData_CountWithTimeData)
	NotifyActivityUpdate_FeedActivityData_GiftBoxData(context.Context, *message.NotifyActivityUpdate_FeedActivityData_GiftBoxData)
	NotifyPayResult_ResourceModify(context.Context, *message.NotifyPayResult_ResourceModify)
	NotifyGameFinishRewardV2_LevelChange(context.Context, *message.NotifyGameFinishRewardV2_LevelChange)
	NotifyGameFinishRewardV2_MatchChest(context.Context, *message.NotifyGameFinishRewardV2_MatchChest)
	NotifyGameFinishRewardV2_MainCharacter(context.Context, *message.NotifyGameFinishRewardV2_MainCharacter)
	NotifyGameFinishRewardV2_CharacterGift(context.Context, *message.NotifyGameFinishRewardV2_CharacterGift)
	NotifyActivityRewardV2_ActivityReward(context.Context, *message.NotifyActivityRewardV2_ActivityReward)
	NotifyActivityPointV2_ActivityPoint(context.Context, *message.NotifyActivityPointV2_ActivityPoint)
	NotifyLeaderboardPointV2_LeaderboardPoint(context.Context, *message.NotifyLeaderboardPointV2_LeaderboardPoint)
	NotifyGameFinishReward_LevelChange(context.Context, *message.NotifyGameFinishReward_LevelChange)
	NotifyGameFinishReward_MatchChest(context.Context, *message.NotifyGameFinishReward_MatchChest)
	NotifyGameFinishReward_MainCharacter(context.Context, *message.NotifyGameFinishReward_MainCharacter)
	NotifyGameFinishReward_CharacterGift(context.Context, *message.NotifyGameFinishReward_CharacterGift)
	NotifyActivityReward_ActivityReward(context.Context, *message.NotifyActivityReward_ActivityReward)
	NotifyActivityPoint_ActivityPoint(context.Context, *message.NotifyActivityPoint_ActivityPoint)
	NotifyLeaderboardPoint_LeaderboardPoint(context.Context, *message.NotifyLeaderboardPoint_LeaderboardPoint)
	NotifyEndGameVote_VoteResult(context.Context, *message.NotifyEndGameVote_VoteResult)
	PlayerLeaving(context.Context, *message.PlayerLeaving)
	ActionPrototype(context.Context, *message.ActionPrototype)
}

func (majsoul *Majsoul) NotifyCaptcha(ctx context.Context, notify *message.NotifyCaptcha) {
}

func (majsoul *Majsoul) NotifyRoomGameStart(ctx context.Context, notify *message.NotifyRoomGameStart) {
	majsoul.ConnGame(ctx)
	var err error
	majsoul.GameInfo, err = majsoul.AuthGame(ctx, &message.ReqAuthGame{
		AccountId: majsoul.Account.AccountId,
		Token:     notify.ConnectToken,
		GameUuid:  notify.GameUuid,
	})
	if err != nil {
		logger.Error("majsoul NotifyRoomGameStart AuthGame error: ", zap.Error(err))
		return
	}

	majsoul.connectToken = notify.ConnectToken
	majsoul.gameUuid = notify.GameUuid

	_, err = majsoul.EnterGame(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul NotifyRoomGameStart EnterGame error:", zap.Error(err))
		return
	}
}

func (majsoul *Majsoul) NotifyMatchGameStart(ctx context.Context, notify *message.NotifyMatchGameStart) {
}

func (majsoul *Majsoul) NotifyRoomPlayerReady(ctx context.Context, notify *message.NotifyRoomPlayerReady) {
}

func (majsoul *Majsoul) NotifyRoomPlayerDressing(ctx context.Context, notify *message.NotifyRoomPlayerDressing) {
}

func (majsoul *Majsoul) NotifyRoomPlayerUpdate(ctx context.Context, notify *message.NotifyRoomPlayerUpdate) {
}

func (majsoul *Majsoul) NotifyRoomKickOut(ctx context.Context, notify *message.NotifyRoomKickOut) {
}

func (majsoul *Majsoul) NotifyFriendStateChange(ctx context.Context, notify *message.NotifyFriendStateChange) {
}

func (majsoul *Majsoul) NotifyFriendViewChange(ctx context.Context, notify *message.NotifyFriendViewChange) {
}

func (majsoul *Majsoul) NotifyFriendChange(ctx context.Context, notify *message.NotifyFriendChange) {
}

func (majsoul *Majsoul) NotifyNewFriendApply(ctx context.Context, notify *message.NotifyNewFriendApply) {
}

func (majsoul *Majsoul) NotifyClientMessage(ctx context.Context, notify *message.NotifyClientMessage) {
}

func (majsoul *Majsoul) NotifyAccountUpdate(ctx context.Context, notify *message.NotifyAccountUpdate) {
}

func (majsoul *Majsoul) NotifyAnotherLogin(ctx context.Context, notify *message.NotifyAnotherLogin) {
}

func (majsoul *Majsoul) NotifyAccountLogout(ctx context.Context, notify *message.NotifyAccountLogout) {
}

func (majsoul *Majsoul) NotifyAnnouncementUpdate(ctx context.Context, notify *message.NotifyAnnouncementUpdate) {
}

func (majsoul *Majsoul) NotifyNewMail(ctx context.Context, notify *message.NotifyNewMail) {
}
func (majsoul *Majsoul) NotifyDeleteMail(ctx context.Context, notify *message.NotifyDeleteMail) {
}
func (majsoul *Majsoul) NotifyReviveCoinUpdate(ctx context.Context, notify *message.NotifyReviveCoinUpdate) {
}

func (majsoul *Majsoul) NotifyDailyTaskUpdate(ctx context.Context, notify *message.NotifyDailyTaskUpdate) {
}

func (majsoul *Majsoul) NotifyActivityTaskUpdate(ctx context.Context, notify *message.NotifyActivityTaskUpdate) {
}

func (majsoul *Majsoul) NotifyActivityPeriodTaskUpdate(ctx context.Context, notify *message.NotifyActivityPeriodTaskUpdate) {
}

func (majsoul *Majsoul) NotifyAccountRandomTaskUpdate(ctx context.Context, notify *message.NotifyAccountRandomTaskUpdate) {
}

func (majsoul *Majsoul) NotifyActivitySegmentTaskUpdate(ctx context.Context, notify *message.NotifyActivitySegmentTaskUpdate) {
}

func (majsoul *Majsoul) NotifyActivityUpdate(ctx context.Context, notify *message.NotifyActivityUpdate) {
}

func (majsoul *Majsoul) NotifyAccountChallengeTaskUpdate(ctx context.Context, notify *message.NotifyAccountChallengeTaskUpdate) {
}

func (majsoul *Majsoul) NotifyNewComment(ctx context.Context, notify *message.NotifyNewComment) {
}

func (majsoul *Majsoul) NotifyRollingNotice(ctx context.Context, notify *message.NotifyRollingNotice) {
}

func (majsoul *Majsoul) NotifyGiftSendRefresh(ctx context.Context, notify *message.NotifyGiftSendRefresh) {
}

func (majsoul *Majsoul) NotifyShopUpdate(ctx context.Context, notify *message.NotifyShopUpdate) {
}

func (majsoul *Majsoul) NotifyVipLevelChange(ctx context.Context, notify *message.NotifyVipLevelChange) {
}

func (majsoul *Majsoul) NotifyServerSetting(ctx context.Context, notify *message.NotifyServerSetting) {
}

func (majsoul *Majsoul) NotifyPayResult(ctx context.Context, notify *message.NotifyPayResult) {
}

func (majsoul *Majsoul) NotifyCustomContestAccountMsg(ctx context.Context, notify *message.NotifyCustomContestAccountMsg) {
}

func (majsoul *Majsoul) NotifyCustomContestSystemMsg(ctx context.Context, notify *message.NotifyCustomContestSystemMsg) {
}

func (majsoul *Majsoul) NotifyMatchTimeout(ctx context.Context, notify *message.NotifyMatchTimeout) {
}

func (majsoul *Majsoul) NotifyCustomContestState(ctx context.Context, notify *message.NotifyCustomContestState) {
}

func (majsoul *Majsoul) NotifyActivityChange(ctx context.Context, notify *message.NotifyActivityChange) {
}

func (majsoul *Majsoul) NotifyAFKResult(ctx context.Context, notify *message.NotifyAFKResult) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2(ctx context.Context, notify *message.NotifyGameFinishRewardV2) {
}

func (majsoul *Majsoul) NotifyActivityRewardV2(ctx context.Context, notify *message.NotifyActivityRewardV2) {
}

func (majsoul *Majsoul) NotifyActivityPointV2(ctx context.Context, notify *message.NotifyActivityPointV2) {
}

func (majsoul *Majsoul) NotifyLeaderboardPointV2(ctx context.Context, notify *message.NotifyLeaderboardPointV2) {
}

func (majsoul *Majsoul) NotifyNewGame(ctx context.Context, notify *message.NotifyNewGame) {
}

func (majsoul *Majsoul) NotifyPlayerLoadGameReady(ctx context.Context, notify *message.NotifyPlayerLoadGameReady) {
}

func (majsoul *Majsoul) NotifyGameBroadcast(ctx context.Context, notify *message.NotifyGameBroadcast) {
}

func (majsoul *Majsoul) NotifyGameEndResult(ctx context.Context, notify *message.NotifyGameEndResult) {
	majsoul.closeFastTestClient()
	majsoul.connectToken = ""
	majsoul.gameUuid = ""
}

func (majsoul *Majsoul) NotifyGameTerminate(ctx context.Context, notify *message.NotifyGameTerminate) {
	majsoul.closeFastTestClient()
	majsoul.connectToken = ""
	majsoul.gameUuid = ""
}

func (majsoul *Majsoul) NotifyPlayerConnectionState(ctx context.Context, notify *message.NotifyPlayerConnectionState) {
}

func (majsoul *Majsoul) NotifyAccountLevelChange(ctx context.Context, notify *message.NotifyAccountLevelChange) {
}

func (majsoul *Majsoul) NotifyGameFinishReward(ctx context.Context, notify *message.NotifyGameFinishReward) {
}

func (majsoul *Majsoul) NotifyActivityReward(ctx context.Context, notify *message.NotifyActivityReward) {
}

func (majsoul *Majsoul) NotifyActivityPoint(ctx context.Context, notify *message.NotifyActivityPoint) {
}

func (majsoul *Majsoul) NotifyLeaderboardPoint(ctx context.Context, notify *message.NotifyLeaderboardPoint) {
}

func (majsoul *Majsoul) NotifyGamePause(ctx context.Context, notify *message.NotifyGamePause) {
}

func (majsoul *Majsoul) NotifyEndGameVote(ctx context.Context, notify *message.NotifyEndGameVote) {
}

func (majsoul *Majsoul) NotifyObserveData(ctx context.Context, notify *message.NotifyObserveData) {
}

func (majsoul *Majsoul) NotifyRoomPlayerReady_AccountReadyState(ctx context.Context, notify *message.NotifyRoomPlayerReady_AccountReadyState) {
}

func (majsoul *Majsoul) NotifyRoomPlayerDressing_AccountDressingState(ctx context.Context, notify *message.NotifyRoomPlayerDressing_AccountDressingState) {
}

func (majsoul *Majsoul) NotifyAnnouncementUpdate_AnnouncementUpdate(ctx context.Context, notify *message.NotifyAnnouncementUpdate_AnnouncementUpdate) {
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData(ctx context.Context, notify *message.NotifyActivityUpdate_FeedActivityData) {
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData_CountWithTimeData(ctx context.Context, notify *message.NotifyActivityUpdate_FeedActivityData_CountWithTimeData) {
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData_GiftBoxData(ctx context.Context, notify *message.NotifyActivityUpdate_FeedActivityData_GiftBoxData) {
}

func (majsoul *Majsoul) NotifyPayResult_ResourceModify(ctx context.Context, notify *message.NotifyPayResult_ResourceModify) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_LevelChange(ctx context.Context, notify *message.NotifyGameFinishRewardV2_LevelChange) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_MatchChest(ctx context.Context, notify *message.NotifyGameFinishRewardV2_MatchChest) {

}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_MainCharacter(ctx context.Context, notify *message.NotifyGameFinishRewardV2_MainCharacter) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_CharacterGift(ctx context.Context, notify *message.NotifyGameFinishRewardV2_CharacterGift) {
}

func (majsoul *Majsoul) NotifyActivityRewardV2_ActivityReward(ctx context.Context, notify *message.NotifyActivityRewardV2_ActivityReward) {
}

func (majsoul *Majsoul) NotifyActivityPointV2_ActivityPoint(ctx context.Context, notify *message.NotifyActivityPointV2_ActivityPoint) {
}

func (majsoul *Majsoul) NotifyLeaderboardPointV2_LeaderboardPoint(ctx context.Context, notify *message.NotifyLeaderboardPointV2_LeaderboardPoint) {

}

func (majsoul *Majsoul) NotifyGameFinishReward_LevelChange(ctx context.Context, notify *message.NotifyGameFinishReward_LevelChange) {

}

func (majsoul *Majsoul) NotifyGameFinishReward_MatchChest(ctx context.Context, notify *message.NotifyGameFinishReward_MatchChest) {

}

func (majsoul *Majsoul) NotifyGameFinishReward_MainCharacter(ctx context.Context, notify *message.NotifyGameFinishReward_MainCharacter) {

}

func (majsoul *Majsoul) NotifyGameFinishReward_CharacterGift(ctx context.Context, notify *message.NotifyGameFinishReward_CharacterGift) {

}

func (majsoul *Majsoul) NotifyActivityReward_ActivityReward(ctx context.Context, notify *message.NotifyActivityReward_ActivityReward) {

}

func (majsoul *Majsoul) NotifyActivityPoint_ActivityPoint(ctx context.Context, notify *message.NotifyActivityPoint_ActivityPoint) {

}

func (majsoul *Majsoul) NotifyLeaderboardPoint_LeaderboardPoint(ctx context.Context, notify *message.NotifyLeaderboardPoint_LeaderboardPoint) {

}

func (majsoul *Majsoul) NotifyEndGameVote_VoteResult(ctx context.Context, notify *message.NotifyEndGameVote_VoteResult) {

}

func (majsoul *Majsoul) PlayerLeaving(ctx context.Context, notify *message.PlayerLeaving) {
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
	actionMessage := message.GetActionType(notify.Name)
	deData := decode(notify.Data)
	err := proto.Unmarshal(deData, actionMessage)
	if err != nil {
		logger.Error("ActionPrototype Unmarshal notify data failed: ", zap.Error(err), zap.Reflect("notify", notify), zap.Binary("data", notify.Data), zap.Binary("deCode", deData))
		return
	}
	switch notify.Name {
	case "ActionMJStart":
		logger.Debug("majsoul ActionMJStart.", zap.Reflect("message", actionMessage))
		majsoul.ActionMJStart(ctx, actionMessage.(*message.ActionMJStart))
		majsoul.implement.ActionMJStart(ctx, actionMessage.(*message.ActionMJStart))
	case "ActionNewCard":
		logger.Debug("majsoul ActionNewCard.", zap.Reflect("message", actionMessage))
		majsoul.ActionNewCard(ctx, actionMessage.(*message.ActionNewCard))
		majsoul.implement.ActionNewCard(ctx, actionMessage.(*message.ActionNewCard))
	case "ActionNewRound":
		logger.Debug("majsoul ActionNewRound.", zap.Reflect("message", actionMessage))
		majsoul.ActionNewRound(ctx, actionMessage.(*message.ActionNewRound))
		majsoul.implement.ActionNewRound(ctx, actionMessage.(*message.ActionNewRound))
	case "ActionSelectGap":
		logger.Debug("majsoul ActionSelectGap.", zap.Reflect("message", actionMessage))
		majsoul.ActionSelectGap(ctx, actionMessage.(*message.ActionSelectGap))
		majsoul.implement.ActionSelectGap(ctx, actionMessage.(*message.ActionSelectGap))
	case "ActionChangeTile":
		logger.Debug("majsoul ActionChangeTile.", zap.Reflect("message", actionMessage))
		majsoul.ActionChangeTile(ctx, actionMessage.(*message.ActionChangeTile))
		majsoul.implement.ActionChangeTile(ctx, actionMessage.(*message.ActionChangeTile))
	case "ActionRevealTile":
		logger.Debug("majsoul ActionRevealTile.", zap.Reflect("message", actionMessage))
		majsoul.ActionRevealTile(ctx, actionMessage.(*message.ActionRevealTile))
		majsoul.implement.ActionRevealTile(ctx, actionMessage.(*message.ActionRevealTile))
	case "ActionUnveilTile":
		logger.Debug("majsoul ActionUnveilTile.", zap.Reflect("message", actionMessage))
		majsoul.ActionUnveilTile(ctx, actionMessage.(*message.ActionUnveilTile))
		majsoul.implement.ActionUnveilTile(ctx, actionMessage.(*message.ActionUnveilTile))
	case "ActionLockTile":
		logger.Debug("majsoul ActionLockTile.", zap.Reflect("message", actionMessage))
		majsoul.ActionLockTile(ctx, actionMessage.(*message.ActionLockTile))
		majsoul.implement.ActionLockTile(ctx, actionMessage.(*message.ActionLockTile))
	case "ActionDiscardTile":
		logger.Debug("majsoul ActionDiscardTile.", zap.Reflect("message", actionMessage))
		majsoul.ActionDiscardTile(ctx, actionMessage.(*message.ActionDiscardTile))
		majsoul.implement.ActionDiscardTile(ctx, actionMessage.(*message.ActionDiscardTile))
	case "ActionDealTile":
		logger.Debug("majsoul ActionDealTile.", zap.Reflect("message", actionMessage))
		majsoul.ActionDealTile(ctx, actionMessage.(*message.ActionDealTile))
		majsoul.implement.ActionDealTile(ctx, actionMessage.(*message.ActionDealTile))
	case "ActionChiPengGang":
		logger.Debug("majsoul ActionChiPengGang.", zap.Reflect("message", actionMessage))
		majsoul.ActionChiPengGang(ctx, actionMessage.(*message.ActionChiPengGang))
		majsoul.implement.ActionChiPengGang(ctx, actionMessage.(*message.ActionChiPengGang))
	case "ActionGangResult":
		logger.Debug("majsoul ActionGangResult.", zap.Reflect("message", actionMessage))
		majsoul.ActionGangResult(ctx, actionMessage.(*message.ActionGangResult))
		majsoul.implement.ActionGangResult(ctx, actionMessage.(*message.ActionGangResult))
	case "ActionGangResultEnd":
		logger.Debug("majsoul ActionGangResultEnd.", zap.Reflect("message", actionMessage))
		majsoul.ActionGangResultEnd(ctx, actionMessage.(*message.ActionGangResultEnd))
		majsoul.implement.ActionGangResultEnd(ctx, actionMessage.(*message.ActionGangResultEnd))
	case "ActionAnGangAddGang":
		logger.Debug("majsoul ActionAnGangAddGang.", zap.Reflect("message", actionMessage))
		majsoul.ActionAnGangAddGang(ctx, actionMessage.(*message.ActionAnGangAddGang))
		majsoul.implement.ActionAnGangAddGang(ctx, actionMessage.(*message.ActionAnGangAddGang))
	case "ActionBaBei":
		logger.Debug("majsoul ActionBaBei.", zap.Reflect("message", actionMessage))
		majsoul.ActionBaBei(ctx, actionMessage.(*message.ActionBaBei))
		majsoul.implement.ActionBaBei(ctx, actionMessage.(*message.ActionBaBei))
	case "ActionHule":
		logger.Debug("majsoul ActionHule.", zap.Reflect("message", actionMessage))
		majsoul.ActionHule(ctx, actionMessage.(*message.ActionHule))
		majsoul.implement.ActionHule(ctx, actionMessage.(*message.ActionHule))
	case "ActionHuleXueZhanMid":
		logger.Debug("majsoul ActionHuleXueZhanMid.", zap.Reflect("message", actionMessage))
		majsoul.ActionHuleXueZhanMid(ctx, actionMessage.(*message.ActionHuleXueZhanMid))
		majsoul.implement.ActionHuleXueZhanMid(ctx, actionMessage.(*message.ActionHuleXueZhanMid))
	case "ActionHuleXueZhanEnd":
		logger.Debug("majsoul ActionHuleXueZhanEnd.", zap.Reflect("message", actionMessage))
		majsoul.ActionHuleXueZhanEnd(ctx, actionMessage.(*message.ActionHuleXueZhanEnd))
		majsoul.implement.ActionHuleXueZhanEnd(ctx, actionMessage.(*message.ActionHuleXueZhanEnd))
	case "ActionLiuJu":
		logger.Debug("majsoul ActionLiuJu.", zap.Reflect("message", actionMessage))
		majsoul.ActionLiuJu(ctx, actionMessage.(*message.ActionLiuJu))
		majsoul.implement.ActionLiuJu(ctx, actionMessage.(*message.ActionLiuJu))
	case "ActionNoTile":
		logger.Debug("majsoul ActionNoTile.", zap.Reflect("message", actionMessage))
		majsoul.ActionNoTile(ctx, actionMessage.(*message.ActionNoTile))
		majsoul.implement.ActionNoTile(ctx, actionMessage.(*message.ActionNoTile))
	default:
		logger.Error("majsoul unknown action prototype name: ", zap.String("name", notify.Name))
	}
}
