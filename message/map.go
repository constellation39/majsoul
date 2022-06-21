package message

import (
	"github.com/golang/protobuf/proto"
	"log"
)

// GetNotifyType 通过api名字返回具体的结构类型
func GetNotifyType(name string) (ret proto.Message) {
	switch name {
	case ".lq.NotifyCaptcha":
		ret = &NotifyCaptcha{}
	case ".lq.NotifyRoomGameStart":
		ret = &NotifyRoomGameStart{}
	case ".lq.NotifyMatchGameStart":
		ret = &NotifyMatchGameStart{}
	case ".lq.NotifyRoomPlayerReady":
		ret = &NotifyRoomPlayerReady{}
	case ".lq.NotifyRoomPlayerDressing":
		ret = &NotifyRoomPlayerDressing{}
	case ".lq.NotifyRoomPlayerUpdate":
		ret = &NotifyRoomPlayerUpdate{}
	case ".lq.NotifyRoomKickOut":
		ret = &NotifyRoomKickOut{}
	case ".lq.NotifyFriendStateChange":
		ret = &NotifyFriendStateChange{}
	case ".lq.NotifyFriendViewChange":
		ret = &NotifyFriendViewChange{}
	case ".lq.NotifyFriendChange":
		ret = &NotifyFriendChange{}
	case ".lq.NotifyNewFriendApply":
		ret = &NotifyNewFriendApply{}
	case ".lq.NotifyClientMessage":
		ret = &NotifyClientMessage{}
	case ".lq.NotifyAccountUpdate":
		ret = &NotifyAccountUpdate{}
	case ".lq.NotifyAnotherLogin":
		ret = &NotifyAnotherLogin{}
	case ".lq.NotifyAccountLogout":
		ret = &NotifyAccountLogout{}
	case ".lq.NotifyAnnouncementUpdate":
		ret = &NotifyAnnouncementUpdate{}
	case ".lq.NotifyNewMail":
		ret = &NotifyNewMail{}
	case ".lq.NotifyDeleteMail":
		ret = &NotifyDeleteMail{}
	case ".lq.NotifyReviveCoinUpdate":
		ret = &NotifyReviveCoinUpdate{}
	case ".lq.NotifyDailyTaskUpdate":
		ret = &NotifyDailyTaskUpdate{}
	case ".lq.NotifyActivityTaskUpdate":
		ret = &NotifyActivityTaskUpdate{}
	case ".lq.NotifyActivityPeriodTaskUpdate":
		ret = &NotifyActivityPeriodTaskUpdate{}
	case ".lq.NotifyAccountRandomTaskUpdate":
		ret = &NotifyAccountRandomTaskUpdate{}
	case ".lq.NotifyActivitySegmentTaskUpdate":
		ret = &NotifyActivitySegmentTaskUpdate{}
	case ".lq.NotifyActivityUpdate":
		ret = &NotifyActivityUpdate{}
	case ".lq.NotifyAccountChallengeTaskUpdate":
		ret = &NotifyAccountChallengeTaskUpdate{}
	case ".lq.NotifyNewComment":
		ret = &NotifyNewComment{}
	case ".lq.NotifyRollingNotice":
		ret = &NotifyRollingNotice{}
	case ".lq.NotifyGiftSendRefresh":
		ret = &NotifyGiftSendRefresh{}
	case ".lq.NotifyShopUpdate":
		ret = &NotifyShopUpdate{}
	case ".lq.NotifyVipLevelChange":
		ret = &NotifyVipLevelChange{}
	case ".lq.NotifyServerSetting":
		ret = &NotifyServerSetting{}
	case ".lq.NotifyPayResult":
		ret = &NotifyPayResult{}
	case ".lq.NotifyCustomContestAccountMsg":
		ret = &NotifyCustomContestAccountMsg{}
	case ".lq.NotifyCustomContestSystemMsg":
		ret = &NotifyCustomContestSystemMsg{}
	case ".lq.NotifyMatchTimeout":
		ret = &NotifyMatchTimeout{}
	case ".lq.NotifyCustomContestState":
		ret = &NotifyCustomContestState{}
	case ".lq.NotifyActivityChange":
		ret = &NotifyActivityChange{}
	case ".lq.NotifyAFKResult":
		ret = &NotifyAFKResult{}
	case ".lq.NotifyGameFinishRewardV2":
		ret = &NotifyGameFinishRewardV2{}
	case ".lq.NotifyActivityRewardV2":
		ret = &NotifyActivityRewardV2{}
	case ".lq.NotifyActivityPointV2":
		ret = &NotifyActivityPointV2{}
	case ".lq.NotifyLeaderboardPointV2":
		ret = &NotifyLeaderboardPointV2{}
	case ".lq.NotifyNewGame":
		ret = &NotifyNewGame{}
	case ".lq.NotifyPlayerLoadGameReady":
		ret = &NotifyPlayerLoadGameReady{}
	case ".lq.NotifyGameBroadcast":
		ret = &NotifyGameBroadcast{}
	case ".lq.NotifyGameEndResult":
		ret = &NotifyGameEndResult{}
	case ".lq.NotifyGameTerminate":
		ret = &NotifyGameTerminate{}
	case ".lq.NotifyPlayerConnectionState":
		ret = &NotifyPlayerConnectionState{}
	case ".lq.NotifyAccountLevelChange":
		ret = &NotifyAccountLevelChange{}
	case ".lq.NotifyGameFinishReward":
		ret = &NotifyGameFinishReward{}
	case ".lq.NotifyActivityReward":
		ret = &NotifyActivityReward{}
	case ".lq.NotifyActivityPoint":
		ret = &NotifyActivityPoint{}
	case ".lq.NotifyLeaderboardPoint":
		ret = &NotifyLeaderboardPoint{}
	case ".lq.NotifyGamePause":
		ret = &NotifyGamePause{}
	case ".lq.NotifyEndGameVote":
		ret = &NotifyEndGameVote{}
	case ".lq.NotifyObserveData":
		ret = &NotifyObserveData{}
	case ".lq.NotifyRoomPlayerReady_AccountReadyState":
		ret = &NotifyRoomPlayerReady_AccountReadyState{}
	case ".lq.NotifyRoomPlayerDressing_AccountDressingState":
		ret = &NotifyRoomPlayerDressing_AccountDressingState{}
	case ".lq.NotifyAnnouncementUpdate_AnnouncementUpdate":
		ret = &NotifyAnnouncementUpdate_AnnouncementUpdate{}
	case ".lq.NotifyActivityUpdate_FeedActivityData":
		ret = &NotifyActivityUpdate_FeedActivityData{}
	case ".lq.NotifyActivityUpdate_FeedActivityData_CountWithTimeData":
		ret = &NotifyActivityUpdate_FeedActivityData_CountWithTimeData{}
	case ".lq.NotifyActivityUpdate_FeedActivityData_GiftBoxData":
		ret = &NotifyActivityUpdate_FeedActivityData_GiftBoxData{}
	case ".lq.NotifyPayResult_ResourceModify":
		ret = &NotifyPayResult_ResourceModify{}
	case ".lq.NotifyGameFinishRewardV2_LevelChange":
		ret = &NotifyGameFinishRewardV2_LevelChange{}
	case ".lq.NotifyGameFinishRewardV2_MatchChest":
		ret = &NotifyGameFinishRewardV2_MatchChest{}
	case ".lq.NotifyGameFinishRewardV2_MainCharacter":
		ret = &NotifyGameFinishRewardV2_MainCharacter{}
	case ".lq.NotifyGameFinishRewardV2_CharacterGift":
		ret = &NotifyGameFinishRewardV2_CharacterGift{}
	case ".lq.NotifyActivityRewardV2_ActivityReward":
		ret = &NotifyActivityRewardV2_ActivityReward{}
	case ".lq.NotifyActivityPointV2_ActivityPoint":
		ret = &NotifyActivityPointV2_ActivityPoint{}
	case ".lq.NotifyLeaderboardPointV2_LeaderboardPoint":
		ret = &NotifyLeaderboardPointV2_LeaderboardPoint{}
	case ".lq.NotifyGameFinishReward_LevelChange":
		ret = &NotifyGameFinishReward_LevelChange{}
	case ".lq.NotifyGameFinishReward_MatchChest":
		ret = &NotifyGameFinishReward_MatchChest{}
	case ".lq.NotifyGameFinishReward_MainCharacter":
		ret = &NotifyGameFinishReward_MainCharacter{}
	case ".lq.NotifyGameFinishReward_CharacterGift":
		ret = &NotifyGameFinishReward_CharacterGift{}
	case ".lq.NotifyActivityReward_ActivityReward":
		ret = &NotifyActivityReward_ActivityReward{}
	case ".lq.NotifyActivityPoint_ActivityPoint":
		ret = &NotifyActivityPoint_ActivityPoint{}
	case ".lq.NotifyLeaderboardPoint_LeaderboardPoint":
		ret = &NotifyLeaderboardPoint_LeaderboardPoint{}
	case ".lq.NotifyEndGameVote_VoteResult":
		ret = &NotifyEndGameVote_VoteResult{}
	case ".lq.PlayerLeaving":
		ret = &PlayerLeaving{}
	case ".lq.ActionPrototype":
		ret = &ActionPrototype{}
	default:
		log.Printf("message.GetNotifyType unknown message type: %s \n", name)
	}
	return
}

// GetActionType 通过api名字返回具体的结构体类型
func GetActionType(name string) (ret proto.Message) {
	switch name {
	case "ActionMJStart":
		ret = &ActionMJStart{}
	case "ActionNewCard":
		ret = &ActionNewCard{}
	case "ActionNewRound":
		ret = &ActionNewRound{}
	case "ActionPrototype":
		ret = &ActionPrototype{}
	case "ActionSelectGap":
		ret = &ActionSelectGap{}
	case "ActionChangeTile":
		ret = &ActionChangeTile{}
	case "ActionRevealTile":
		ret = &ActionRevealTile{}
	case "ActionUnveilTile":
		ret = &ActionUnveilTile{}
	case "ActionLockTile":
		ret = &ActionLockTile{}
	case "ActionDiscardTile":
		ret = &ActionDiscardTile{}
	case "ActionDealTile":
		ret = &ActionDealTile{}
	case "ActionChiPengGang":
		ret = &ActionChiPengGang{}
	case "ActionGangResult":
		ret = &ActionGangResult{}
	case "ActionGangResultEnd":
		ret = &ActionGangResultEnd{}
	case "ActionAnGangAddGang":
		ret = &ActionAnGangAddGang{}
	case "ActionBaBei":
		ret = &ActionBaBei{}
	case "ActionHule":
		ret = &ActionHule{}
	case "ActionHuleXueZhanMid":
		ret = &ActionHuleXueZhanMid{}
	case "ActionHuleXueZhanEnd":
		ret = &ActionHuleXueZhanEnd{}
	case "ActionLiuJu":
		ret = &ActionLiuJu{}
	case "ActionNoTile":
		ret = &ActionNoTile{}
	default:
		log.Printf("message.GetActionType unknown message type: %s \n", name)
	}
	return
}
