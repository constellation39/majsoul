package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
	"time"
)

type GameState struct {
	seat         uint32
	account      *message.Account     // 该字段应在登录成功后访问
	gameInfo     *message.ResAuthGame // 该字段应在进入游戏桌面后访问
	accessToken  string               // 验证身份时使用 的 token
	connectToken string               // 重连时使用的 token
	gameUuid     string               // 是否在游戏中
}

func (gameState *GameState) NotifyClientMessage(majSoul *majsoul.MajSoul, notifyClientMessage *message.NotifyClientMessage) {
	type DetailRule struct {
		TimeFixed    int  `json:"time_fixed"`
		TimeAdd      int  `json:"time_add"`
		DoraCount    int  `json:"dora_count"`
		Shiduan      int  `json:"shiduan"`
		InitPoint    int  `json:"init_point"`
		Fandian      int  `json@:"fandian"`
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
	// 我们现在只处理 type == 1 , 也就是收到邀请的情况
	if notifyClientMessage.Type != 1 {
		logger.Info("notifyClientMessage.Type != -1", zap.Uint32("type", notifyClientMessage.Type))
		return
	}
	invitationRoom := new(InvitationRoom)
	err := json.Unmarshal([]byte(notifyClientMessage.Content), invitationRoom)
	if err != nil {
		logger.Error("Unmarshal", zap.Error(err))
		return
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		// 加入房间
		_, err = majSoul.LobbyClient.JoinRoom(ctx, &message.ReqJoinRoom{
			RoomId:              invitationRoom.RoomID,
			ClientVersionString: majSoul.Version.Web(),
		})
		if err != nil {
			logger.Error("JoinRoom", zap.Error(err))
			return
		}
	}

	time.Sleep(time.Second)

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		// 准备
		_, err = majSoul.LobbyClient.ReadyPlay(ctx, &message.ReqRoomReady{Ready: true})
		if err != nil {
			logger.Error("ReadyPlay", zap.Error(err))
			return
		}
	}
}

func (gameState *GameState) NotifyFriendViewChange(majSoul *majsoul.MajSoul, notifyFriendViewChange *message.NotifyFriendViewChange) {
	logger.Debug("", zap.Reflect("notifyFriendViewChange", notifyFriendViewChange))
}

// NotifyEndGameVote 有人发起投降
func (gameState *GameState) NotifyEndGameVote(majSoul *majsoul.MajSoul, notifyEndGameVote *message.NotifyEndGameVote) {
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.FastTestClient.VoteGameEnd(ctx, &message.ReqVoteGameEnd{Yes: true})
		if err != nil {
			logger.Error("VoteGameEnd", zap.Error(err))
		}
	}
}

// 从等待房间进入游戏时
func (gameState *GameState) NotifyRoomGameStart(majSoul *majsoul.MajSoul, notifyRoomGameStart *message.NotifyRoomGameStart) {

	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := majSoul.ConnGame(ctx)
		if err != nil {
			panic(fmt.Sprintf("conn GameState server failed error %v", err))
		}
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		var err error
		gameState.gameInfo, err = majSoul.FastTestClient.AuthGame(ctx, &message.ReqAuthGame{
			AccountId: gameState.account.AccountId,
			Token:     gameState.connectToken,
			GameUuid:  gameState.gameUuid,
		})
		if err != nil {
			logger.Error("majsoul NotifyRoomGameStart AuthGame error: ", zap.Error(err))
			return
		}
	}
	gameState.connectToken = notifyRoomGameStart.ConnectToken
	gameState.gameUuid = notifyRoomGameStart.GameUuid
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.FastTestClient.EnterGame(ctx, &message.ReqCommon{})
		if err != nil {
			logger.Error("majsoul NotifyRoomGameStart EnterGame error:", zap.Error(err))
			return
		}
	}

	// 记录自己的座位号
	for i, uid := range gameState.gameInfo.SeatList {
		if uid == gameState.account.AccountId {
			gameState.seat = uint32(i)
			break
		}
	}
}

// ActionMJStart 游戏开始
func (gameState *GameState) ActionMJStart(majSoul *majsoul.MajSoul, actionMJStart *message.ActionMJStart) {
}

// ActionNewRound 回合开始
func (gameState *GameState) ActionNewRound(majSoul *majsoul.MajSoul, action *message.ActionNewRound) {
	// 如果是庄家
	if len(action.Tiles) != 14 {
		return
	}
	tile13 := action.Tiles[13]
	time.Sleep(time.Second * 3)
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.FastTestClient.InputOperation(ctx, &message.ReqSelfOperation{
			Type:    majsoul.ActionDiscard,
			Tile:    tile13,
			Moqie:   true,
			Timeuse: 1,
		})
		if err != nil {
			logger.Error("InputOperation failed", zap.Error(err))
		}
	}
}

// ActionDealTile 摸牌
func (gameState *GameState) ActionDealTile(majSoul *majsoul.MajSoul, action *message.ActionDealTile) {
	// 如果不是自己摸牌
	if action.Seat != gameState.seat {
		return
	}

	if len(action.Tile) == 0 {
		logger.Error("摸牌是空的")
		return
	}

	time.Sleep(time.Second * 3)
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := majSoul.FastTestClient.InputOperation(ctx, &message.ReqSelfOperation{
			Type:    majsoul.ActionDiscard,
			Tile:    action.Tile,
			Moqie:   true,
			Timeuse: 1,
		})
		if err != nil {
			logger.Error("InputOperation failed", zap.Error(err))
		}
	}
}

// ActionDiscardTile 打牌
func (gameState *GameState) ActionDiscardTile(majSoul *majsoul.MajSoul, action *message.ActionDiscardTile) {
	if action.Operation != nil && len(action.Operation.OperationList) != 0 {
		for _, operation := range action.Operation.OperationList {
			switch operation.Type {
			case majsoul.ActionDiscard:
			case majsoul.ActionChi:
			case majsoul.ActionPon:
			case majsoul.ActionAnKAN:
			case majsoul.ActionMinKan:
			case majsoul.ActionKaKan:
			case majsoul.ActionRiichi:
			case majsoul.ActionTsumo:
			case majsoul.ActionRon:
			case majsoul.ActionKuku:
			case majsoul.ActionKita:
			case majsoul.ActionPass:
				if action.Operation != nil {
					{
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
						defer cancel()
						_, err := majSoul.FastTestClient.InputOperation(ctx, &message.ReqSelfOperation{
							CancelOperation: true,
							Timeuse:         1,
						})
						if err != nil {
							logger.Error("InputOperation failed", zap.Error(err))
						}
					}
				}
			}
		}
	}
}

// ActionChiPengGang 吃碰杠的通知
func (gameState *GameState) ActionChiPengGang(majSoul *majsoul.MajSoul, action *message.ActionChiPengGang) {
	switch action.Type {
	case majsoul.NotifyChi:
	case majsoul.NotifyPon:
	case majsoul.NotifyKan:
	}
}

// ActionAnGangAddGang 暗杠和加杠的通知
func (gameState *GameState) ActionAnGangAddGang(majSoul *majsoul.MajSoul, action *message.ActionAnGangAddGang) {
	switch action.Type {
	case majsoul.NotifyAnKan:
	case majsoul.NotifyKaKan:
	}
}

func (gameState *GameState) ActionHule(majSoul *majsoul.MajSoul, action *message.ActionHule) {
}

func (gameState *GameState) ActionLiuJu(majSoul *majsoul.MajSoul, action *message.ActionLiuJu) {
}

func (gameState *GameState) ActionNoTile(majSoul *majsoul.MajSoul, action *message.ActionNoTile) {
}
