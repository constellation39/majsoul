package majsoul

import "github.com/constellation39/majsoul/message"

// IFAction 游戏内消息接口
type IFAction interface {
	ActionMJStart(*message.ActionMJStart)
	ActionNewCard(*message.ActionNewCard)
	ActionNewRound(*message.ActionNewRound)
	ActionSelectGap(*message.ActionSelectGap)
	ActionChangeTile(*message.ActionChangeTile)
	ActionRevealTile(*message.ActionRevealTile)
	ActionUnveilTile(*message.ActionUnveilTile)
	ActionLockTile(*message.ActionLockTile)
	ActionDiscardTile(*message.ActionDiscardTile)
	ActionDealTile(*message.ActionDealTile)
	ActionChiPengGang(*message.ActionChiPengGang)
	ActionGangResult(*message.ActionGangResult)
	ActionGangResultEnd(*message.ActionGangResultEnd)
	ActionAnGangAddGang(*message.ActionAnGangAddGang)
	ActionBaBei(*message.ActionBaBei)
	ActionHule(*message.ActionHule)
	ActionHuleXueZhanMid(*message.ActionHuleXueZhanMid)
	ActionHuleXueZhanEnd(*message.ActionHuleXueZhanEnd)
	ActionLiuJu(*message.ActionLiuJu)
	ActionNoTile(*message.ActionNoTile)
}
