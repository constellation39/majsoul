package main

import (
	"context"
	"encoding/json"
	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
	"time"
)

type GameState struct {
}

func (GameState) NotifyClientMessage(majSoul *majsoul.MajSoul, notifyClientMessage *message.NotifyClientMessage) {
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
		_, err = majSoul.JoinRoom(ctx, &message.ReqJoinRoom{
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
		_, err = majSoul.ReadyPlay(ctx, &message.ReqRoomReady{Ready: true})
		if err != nil {
			logger.Error("ReadyPlay", zap.Error(err))
			return
		}
	}
}

func (GameState) NotifyFriendViewChange(majSoul *majsoul.MajSoul, notifyFriendViewChange *message.NotifyFriendViewChange) {
	logger.Debug("", zap.Reflect("notifyFriendViewChange", notifyFriendViewChange))
}

func (GameState) ActionMJStart(majSoul *majsoul.MajSoul, actionMJStart *message.ActionMJStart) {
	logger.Debug("ActionMJStart")
}

func main() {
	sync := logger.Init()
	defer sync()

	majSoul := majsoul.NewMajSoul(&majsoul.Config{ProxyAddress: ""})
	majSoul.LookupGateway(context.Background(), majsoul.ServerAddressList)

	var gameState GameState

	majSoul.Handle(gameState.NotifyClientMessage, gameState.NotifyClientMessage, gameState.ActionMJStart)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := majSoul.Login(ctx, "1601198895@qq.com", "miku39..")
	if err != nil {
		panic(err)
	}
	select {}
}
