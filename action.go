package majsoul

import (
	"context"

	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
)

// 这里处理的是游戏中的动作数据

func (majsoul *Majsoul) ActionMJStart(ctx context.Context, action *message.ActionMJStart) {
	logger.Debug("majsoul ActionMJStart", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionNewCard(ctx context.Context, action *message.ActionNewCard) {
	logger.Debug("majsoul ActionNewCard", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionNewRound(ctx context.Context, action *message.ActionNewRound) {
	logger.Debug("majsoul ActionNewRound", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionSelectGap(ctx context.Context, action *message.ActionSelectGap) {
	logger.Debug("majsoul ActionSelectGap", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionChangeTile(ctx context.Context, action *message.ActionChangeTile) {
	logger.Debug("majsoul ActionChangeTile", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionRevealTile(ctx context.Context, action *message.ActionRevealTile) {
	logger.Debug("majsoul ActionRevealTile", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionUnveilTile(ctx context.Context, action *message.ActionUnveilTile) {
	logger.Debug("majsoul ActionUnveilTile", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionLockTile(ctx context.Context, action *message.ActionLockTile) {
	logger.Debug("majsoul ActionLockTile", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionDiscardTile(ctx context.Context, action *message.ActionDiscardTile) {
	logger.Debug("majsoul ActionDiscardTile", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionDealTile(ctx context.Context, action *message.ActionDealTile) {
	logger.Debug("majsoul ActionDealTile", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionChiPengGang(ctx context.Context, action *message.ActionChiPengGang) {
	logger.Debug("majsoul ActionChiPengGang", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionGangResult(ctx context.Context, action *message.ActionGangResult) {
	logger.Debug("majsoul ActionGangResult", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionGangResultEnd(ctx context.Context, action *message.ActionGangResultEnd) {
	logger.Debug("majsoul ActionGangResultEnd", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionAnGangAddGang(ctx context.Context, action *message.ActionAnGangAddGang) {
	logger.Debug("majsoul ActionAnGangAddGang", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionBaBei(ctx context.Context, action *message.ActionBaBei) {
	logger.Debug("majsoul ActionBaBei", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionHule(ctx context.Context, action *message.ActionHule) {
	logger.Debug("majsoul ActionHule", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionHuleXueZhanMid(ctx context.Context, action *message.ActionHuleXueZhanMid) {
	logger.Debug("majsoul ActionHuleXueZhanMid", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionHuleXueZhanEnd(ctx context.Context, action *message.ActionHuleXueZhanEnd) {
	logger.Debug("majsoul ActionHuleXueZhanEnd", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionLiuJu(ctx context.Context, action *message.ActionLiuJu) {
	logger.Debug("ActionLiuJu", zap.Reflect("action", action))
}

func (majsoul *Majsoul) ActionNoTile(ctx context.Context, action *message.ActionNoTile) {
	logger.Debug("ActionNoTile", zap.Reflect("action", action))
}
