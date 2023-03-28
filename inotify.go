package majsoul

import (
	"context"

	"github.com/constellation39/majsoul/message"
)

// IFNotify
// 雀魂proto协议中缺少描述监听消息的接口
// 故添加该接口，可能会丢失某些api
// 有没有更聪明点的办法？
type IFNotify interface {
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
