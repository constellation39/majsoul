package majsoul

import (
	"context"

	"github.com/constellation39/majsoul/message"
)

// Action 游戏内消息接口
type Action interface {
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

func (majsoul *Majsoul) ActionMJStart(ctx context.Context, action *message.ActionMJStart) {}

func (majsoul *Majsoul) ActionNewCard(ctx context.Context, action *message.ActionNewCard) {}

func (majsoul *Majsoul) ActionNewRound(ctx context.Context, action *message.ActionNewRound) {}

func (majsoul *Majsoul) ActionSelectGap(ctx context.Context, action *message.ActionSelectGap) {}

func (majsoul *Majsoul) ActionChangeTile(ctx context.Context, action *message.ActionChangeTile) {}

func (majsoul *Majsoul) ActionRevealTile(ctx context.Context, action *message.ActionRevealTile) {}

func (majsoul *Majsoul) ActionUnveilTile(ctx context.Context, action *message.ActionUnveilTile) {}

func (majsoul *Majsoul) ActionLockTile(ctx context.Context, action *message.ActionLockTile) {}

func (majsoul *Majsoul) ActionDiscardTile(ctx context.Context, action *message.ActionDiscardTile) {}

func (majsoul *Majsoul) ActionDealTile(ctx context.Context, action *message.ActionDealTile) {}

func (majsoul *Majsoul) ActionChiPengGang(ctx context.Context, action *message.ActionChiPengGang) {}

func (majsoul *Majsoul) ActionGangResult(ctx context.Context, action *message.ActionGangResult) {}

func (majsoul *Majsoul) ActionGangResultEnd(ctx context.Context, action *message.ActionGangResultEnd) {
}

func (majsoul *Majsoul) ActionAnGangAddGang(ctx context.Context, action *message.ActionAnGangAddGang) {
}

func (majsoul *Majsoul) ActionBaBei(ctx context.Context, action *message.ActionBaBei) {}

func (majsoul *Majsoul) ActionHule(ctx context.Context, action *message.ActionHule) {}

func (majsoul *Majsoul) ActionHuleXueZhanMid(ctx context.Context, action *message.ActionHuleXueZhanMid) {
}

func (majsoul *Majsoul) ActionHuleXueZhanEnd(ctx context.Context, action *message.ActionHuleXueZhanEnd) {
}

func (majsoul *Majsoul) ActionLiuJu(ctx context.Context, action *message.ActionLiuJu) {}

func (majsoul *Majsoul) ActionNoTile(ctx context.Context, action *message.ActionNoTile) {}
