package majsoul

import (
	"context"

	"github.com/constellation39/majsoul/message"
)

// IFAction 游戏内消息接口
type IFAction interface {
	ActionMJStart(context.Context, *message.ActionMJStart)
	ActionNewCard(context.Context, *message.ActionNewCard)
	ActionNewRound(context.Context, *message.ActionNewRound)
	ActionSelectGap(context.Context, *message.ActionSelectGap)
	ActionChangeTile(context.Context, *message.ActionChangeTile)
	ActionRevealTile(context.Context, *message.ActionRevealTile)
	ActionUnveilTile(context.Context, *message.ActionUnveilTile)
	ActionLockTile(context.Context, *message.ActionLockTile)
	ActionDiscardTile(context.Context, *message.ActionDiscardTile)
	ActionDealTile(context.Context, *message.ActionDealTile)
	ActionChiPengGang(context.Context, *message.ActionChiPengGang)
	ActionGangResult(context.Context, *message.ActionGangResult)
	ActionGangResultEnd(context.Context, *message.ActionGangResultEnd)
	ActionAnGangAddGang(context.Context, *message.ActionAnGangAddGang)
	ActionBaBei(context.Context, *message.ActionBaBei)
	ActionHule(context.Context, *message.ActionHule)
	ActionHuleXueZhanMid(context.Context, *message.ActionHuleXueZhanMid)
	ActionHuleXueZhanEnd(context.Context, *message.ActionHuleXueZhanEnd)
	ActionLiuJu(context.Context, *message.ActionLiuJu)
	ActionNoTile(context.Context, *message.ActionNoTile)
}
