package majsoul

import (
	"log"

	"github.com/constellation39/majsoul/message"
	"github.com/golang/protobuf/proto"
)

// IFNotify is the interface that must be implemented by a receiver.

func (majsoul *Majsoul) NotifyCaptcha(notify *message.NotifyCaptcha) {
}

func (majsoul *Majsoul) NotifyRoomGameStart(notify *message.NotifyRoomGameStart) {
	var err error
	majsoul.FastTestConn, err = NewClientConn(majsoul.Ctx, majsoul.ServerAddress.GameAddress)
	if err != nil {
		log.Fatalf("Majsoul.NotifyRoomGameStart Connect to GameServer failed %s", majsoul.ServerAddress.GatewayAddress)
		return
	}
	majsoul.FastTestClient = message.NewFastTestClient(majsoul.FastTestConn)
	go majsoul.receiveGame()
	majsoul.GameInfo, err = majsoul.AuthGame(majsoul.Ctx, &message.ReqAuthGame{
		AccountId: majsoul.Account.AccountId,
		Token:     notify.ConnectToken,
		GameUuid:  notify.GameUuid,
	})
	if err != nil {
		log.Printf("Majsoul.NotifyRoomGameStart AuthGame error: %v \n", err)
		return
	}
	_, err = majsoul.EnterGame(majsoul.Ctx, &message.ReqCommon{})
	if err != nil {
		log.Printf("Majsoul.NotifyRoomGameStart EnterGame error: %v \n", err)
		return
	}
}

func (majsoul *Majsoul) NotifyMatchGameStart(notify *message.NotifyMatchGameStart) {
}

func (majsoul *Majsoul) NotifyRoomPlayerReady(notify *message.NotifyRoomPlayerReady) {
}

func (majsoul *Majsoul) NotifyRoomPlayerDressing(notify *message.NotifyRoomPlayerDressing) {
}

func (majsoul *Majsoul) NotifyRoomPlayerUpdate(notify *message.NotifyRoomPlayerUpdate) {
}

func (majsoul *Majsoul) NotifyRoomKickOut(notify *message.NotifyRoomKickOut) {
}

func (majsoul *Majsoul) NotifyFriendStateChange(notify *message.NotifyFriendStateChange) {
}

func (majsoul *Majsoul) NotifyFriendViewChange(notify *message.NotifyFriendViewChange) {
}

func (majsoul *Majsoul) NotifyFriendChange(notify *message.NotifyFriendChange) {
}

func (majsoul *Majsoul) NotifyNewFriendApply(notify *message.NotifyNewFriendApply) {
}

func (majsoul *Majsoul) NotifyClientMessage(notify *message.NotifyClientMessage) {
}

func (majsoul *Majsoul) NotifyAccountUpdate(notify *message.NotifyAccountUpdate) {
}

func (majsoul *Majsoul) NotifyAnotherLogin(notify *message.NotifyAnotherLogin) {
}

func (majsoul *Majsoul) NotifyAccountLogout(notify *message.NotifyAccountLogout) {
}

func (majsoul *Majsoul) NotifyAnnouncementUpdate(notify *message.NotifyAnnouncementUpdate) {
}

func (majsoul *Majsoul) NotifyNewMail(notify *message.NotifyNewMail) {
}

func (majsoul *Majsoul) NotifyDeleteMail(notify *message.NotifyDeleteMail) {
}

func (majsoul *Majsoul) NotifyReviveCoinUpdate(notify *message.NotifyReviveCoinUpdate) {
}

func (majsoul *Majsoul) NotifyDailyTaskUpdate(notify *message.NotifyDailyTaskUpdate) {
}

func (majsoul *Majsoul) NotifyActivityTaskUpdate(notify *message.NotifyActivityTaskUpdate) {
}

func (majsoul *Majsoul) NotifyActivityPeriodTaskUpdate(notify *message.NotifyActivityPeriodTaskUpdate) {
}

func (majsoul *Majsoul) NotifyAccountRandomTaskUpdate(notify *message.NotifyAccountRandomTaskUpdate) {
}

func (majsoul *Majsoul) NotifyActivitySegmentTaskUpdate(notify *message.NotifyActivitySegmentTaskUpdate) {
}

func (majsoul *Majsoul) NotifyActivityUpdate(notify *message.NotifyActivityUpdate) {
}

func (majsoul *Majsoul) NotifyAccountChallengeTaskUpdate(notify *message.NotifyAccountChallengeTaskUpdate) {
}

func (majsoul *Majsoul) NotifyNewComment(notify *message.NotifyNewComment) {
}

func (majsoul *Majsoul) NotifyRollingNotice(notify *message.NotifyRollingNotice) {
}

func (majsoul *Majsoul) NotifyGiftSendRefresh(notify *message.NotifyGiftSendRefresh) {
}

func (majsoul *Majsoul) NotifyShopUpdate(notify *message.NotifyShopUpdate) {
}

func (majsoul *Majsoul) NotifyVipLevelChange(notify *message.NotifyVipLevelChange) {
}

func (majsoul *Majsoul) NotifyServerSetting(notify *message.NotifyServerSetting) {
}

func (majsoul *Majsoul) NotifyPayResult(notify *message.NotifyPayResult) {
}

func (majsoul *Majsoul) NotifyCustomContestAccountMsg(notify *message.NotifyCustomContestAccountMsg) {
}

func (majsoul *Majsoul) NotifyCustomContestSystemMsg(notify *message.NotifyCustomContestSystemMsg) {
}

func (majsoul *Majsoul) NotifyMatchTimeout(notify *message.NotifyMatchTimeout) {
}

func (majsoul *Majsoul) NotifyCustomContestState(notify *message.NotifyCustomContestState) {
}

func (majsoul *Majsoul) NotifyActivityChange(notify *message.NotifyActivityChange) {
}

func (majsoul *Majsoul) NotifyAFKResult(notify *message.NotifyAFKResult) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2(notify *message.NotifyGameFinishRewardV2) {
}

func (majsoul *Majsoul) NotifyActivityRewardV2(notify *message.NotifyActivityRewardV2) {
}

func (majsoul *Majsoul) NotifyActivityPointV2(notify *message.NotifyActivityPointV2) {
}

func (majsoul *Majsoul) NotifyLeaderboardPointV2(notify *message.NotifyLeaderboardPointV2) {
}

func (majsoul *Majsoul) NotifyNewGame(notify *message.NotifyNewGame) {
}

func (majsoul *Majsoul) NotifyPlayerLoadGameReady(notify *message.NotifyPlayerLoadGameReady) {
}

func (majsoul *Majsoul) NotifyGameBroadcast(notify *message.NotifyGameBroadcast) {
}

func (majsoul *Majsoul) NotifyGameEndResult(notify *message.NotifyGameEndResult) {
}

func (majsoul *Majsoul) NotifyGameTerminate(notify *message.NotifyGameTerminate) {
	majsoul.FastTestConn = nil
}

func (majsoul *Majsoul) NotifyPlayerConnectionState(notify *message.NotifyPlayerConnectionState) {
}

func (majsoul *Majsoul) NotifyAccountLevelChange(notify *message.NotifyAccountLevelChange) {
}

func (majsoul *Majsoul) NotifyGameFinishReward(notify *message.NotifyGameFinishReward) {
}

func (majsoul *Majsoul) NotifyActivityReward(notify *message.NotifyActivityReward) {
}

func (majsoul *Majsoul) NotifyActivityPoint(notify *message.NotifyActivityPoint) {
}

func (majsoul *Majsoul) NotifyLeaderboardPoint(notify *message.NotifyLeaderboardPoint) {
}

func (majsoul *Majsoul) NotifyGamePause(notify *message.NotifyGamePause) {
}

func (majsoul *Majsoul) NotifyEndGameVote(notify *message.NotifyEndGameVote) {
}

func (majsoul *Majsoul) NotifyObserveData(notify *message.NotifyObserveData) {
}

func (majsoul *Majsoul) NotifyRoomPlayerReady_AccountReadyState(notify *message.NotifyRoomPlayerReady_AccountReadyState) {
}

func (majsoul *Majsoul) NotifyRoomPlayerDressing_AccountDressingState(notify *message.NotifyRoomPlayerDressing_AccountDressingState) {
}

func (majsoul *Majsoul) NotifyAnnouncementUpdate_AnnouncementUpdate(notify *message.NotifyAnnouncementUpdate_AnnouncementUpdate) {
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData(notify *message.NotifyActivityUpdate_FeedActivityData) {
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData_CountWithTimeData(notify *message.NotifyActivityUpdate_FeedActivityData_CountWithTimeData) {
}

func (majsoul *Majsoul) NotifyActivityUpdate_FeedActivityData_GiftBoxData(notify *message.NotifyActivityUpdate_FeedActivityData_GiftBoxData) {
}

func (majsoul *Majsoul) NotifyPayResult_ResourceModify(notify *message.NotifyPayResult_ResourceModify) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_LevelChange(notify *message.NotifyGameFinishRewardV2_LevelChange) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_MatchChest(notify *message.NotifyGameFinishRewardV2_MatchChest) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_MainCharacter(notify *message.NotifyGameFinishRewardV2_MainCharacter) {
}

func (majsoul *Majsoul) NotifyGameFinishRewardV2_CharacterGift(notify *message.NotifyGameFinishRewardV2_CharacterGift) {
}

func (majsoul *Majsoul) NotifyActivityRewardV2_ActivityReward(notify *message.NotifyActivityRewardV2_ActivityReward) {
}

func (majsoul *Majsoul) NotifyActivityPointV2_ActivityPoint(notify *message.NotifyActivityPointV2_ActivityPoint) {
}

func (majsoul *Majsoul) NotifyLeaderboardPointV2_LeaderboardPoint(notify *message.NotifyLeaderboardPointV2_LeaderboardPoint) {
}

func (majsoul *Majsoul) NotifyGameFinishReward_LevelChange(notify *message.NotifyGameFinishReward_LevelChange) {
}

func (majsoul *Majsoul) NotifyGameFinishReward_MatchChest(notify *message.NotifyGameFinishReward_MatchChest) {
}

func (majsoul *Majsoul) NotifyGameFinishReward_MainCharacter(notify *message.NotifyGameFinishReward_MainCharacter) {
}

func (majsoul *Majsoul) NotifyGameFinishReward_CharacterGift(notify *message.NotifyGameFinishReward_CharacterGift) {
}

func (majsoul *Majsoul) NotifyActivityReward_ActivityReward(notify *message.NotifyActivityReward_ActivityReward) {
}

func (majsoul *Majsoul) NotifyActivityPoint_ActivityPoint(notify *message.NotifyActivityPoint_ActivityPoint) {
}

func (majsoul *Majsoul) NotifyLeaderboardPoint_LeaderboardPoint(notify *message.NotifyLeaderboardPoint_LeaderboardPoint) {
}

func (majsoul *Majsoul) NotifyEndGameVote_VoteResult(notify *message.NotifyEndGameVote_VoteResult) {
}

func (majsoul *Majsoul) PlayerLeaving(notify *message.PlayerLeaving) {

}

var keys = []int{0x84, 0x5e, 0x4e, 0x42, 0x39, 0xa2, 0x1f, 0x60, 0x1c}

func decode(data []byte) {
	for i := 0; i < len(data); i++ {
		u := (23 ^ len(data)) + 5*i + keys[i%len(keys)]&255
		data[i] ^= byte(u)
	}
}

func (majsoul *Majsoul) ActionPrototype(notify *message.ActionPrototype) {
	data := message.GetActionType(notify.Name)
	decode(notify.Data)
	err := proto.Unmarshal(notify.Data, data)
	if err != nil {
		log.Printf("ActionPrototype Unmarshal error: %v", err)
		return
	}
	switch notify.Name {
	case "ActionMJStart":
		majsoul.Implement.ActionMJStart(data.(*message.ActionMJStart))
	case "ActionNewCard":
		majsoul.Implement.ActionNewCard(data.(*message.ActionNewCard))
	case "ActionNewRound":
		majsoul.Implement.ActionNewRound(data.(*message.ActionNewRound))
	case "ActionSelectGap":
		majsoul.Implement.ActionSelectGap(data.(*message.ActionSelectGap))
	case "ActionChangeTile":
		majsoul.Implement.ActionChangeTile(data.(*message.ActionChangeTile))
	case "ActionRevealTile":
		majsoul.Implement.ActionRevealTile(data.(*message.ActionRevealTile))
	case "ActionUnveilTile":
		majsoul.Implement.ActionUnveilTile(data.(*message.ActionUnveilTile))
	case "ActionLockTile":
		majsoul.Implement.ActionLockTile(data.(*message.ActionLockTile))
	case "ActionDiscardTile":
		majsoul.Implement.ActionDiscardTile(data.(*message.ActionDiscardTile))
	case "ActionDealTile":
		majsoul.Implement.ActionDealTile(data.(*message.ActionDealTile))
	case "ActionChiPengGang":
		majsoul.Implement.ActionChiPengGang(data.(*message.ActionChiPengGang))
	case "ActionGangResult":
		majsoul.Implement.ActionGangResult(data.(*message.ActionGangResult))
	case "ActionGangResultEnd":
		majsoul.Implement.ActionGangResultEnd(data.(*message.ActionGangResultEnd))
	case "ActionAnGangAddGang":
		majsoul.Implement.ActionAnGangAddGang(data.(*message.ActionAnGangAddGang))
	case "ActionBaBei":
		majsoul.Implement.ActionBaBei(data.(*message.ActionBaBei))
	case "ActionHule":
		majsoul.Implement.ActionHule(data.(*message.ActionHule))
	case "ActionHuleXueZhanMid":
		majsoul.Implement.ActionHuleXueZhanMid(data.(*message.ActionHuleXueZhanMid))
	case "ActionHuleXueZhanEnd":
		majsoul.Implement.ActionHuleXueZhanEnd(data.(*message.ActionHuleXueZhanEnd))
	case "ActionLiuJu":
		majsoul.Implement.ActionLiuJu(data.(*message.ActionLiuJu))
	case "ActionNoTile":
		majsoul.Implement.ActionNoTile(data.(*message.ActionNoTile))
	}
}
