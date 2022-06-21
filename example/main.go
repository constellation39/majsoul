package main

import (
	"encoding/json"
	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/message"
	"log"
)

// Majsoul 组合库中的 Majsoul 结构
type Majsoul struct {
	*majsoul.Majsoul
	tiles []string
}

// NewMajsoul 创建一个 Majsoul 结构
func NewMajsoul() *Majsoul {
	config, err := majsoul.LoadConfig("majsoul.json")
	if err != nil {
		log.Fatal(err)
	}
	mSoul := &Majsoul{Majsoul: majsoul.New(config)}
	mSoul.Implement = mSoul // 使用多态实现，如果调用时没有提供外部实现则调用内部的实现，如果没有给 Implement 赋值，则只会调用内部实现
	return mSoul
}

// NotifyClientMessage 客户端消息
// message.NotifyClientMessage filed Type == 1 时为受到邀请
// note: 这个函数的只实现了接受到邀请的通知
func (mSoul *Majsoul) NotifyClientMessage(notify *message.NotifyClientMessage) {
	type DetailRule struct {
		TimeFixed    int  `json:"time_fixed"`
		TimeAdd      int  `json:"time_add"`
		DoraCount    int  `json:"dora_count"`
		Shiduan      int  `json:"shiduan"`
		InitPoint    int  `json:"init_point"`
		Fandian      int  `json:"fandian"`
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
	if notify.Type != 1 {
		log.Printf("%+v", notify)
		return
	}
	invitationRoom := new(InvitationRoom)
	err := json.Unmarshal([]byte(notify.Content), invitationRoom)
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	_, err = mSoul.JoinRoom(mSoul.Ctx, &message.ReqJoinRoom{
		RoomId:              invitationRoom.RoomID,
		ClientVersionString: mSoul.Version.Web(),
	})
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	_, err = mSoul.ReadyPlay(mSoul.Ctx, &message.ReqRoomReady{Ready: true})
	if err != nil {
		log.Printf("%+v", err)
		return
	}
}

// NotifyEndGameVote 有人发起投降
func (mSoul *Majsoul) NotifyEndGameVote(notify *message.NotifyEndGameVote) {
	end, err := mSoul.VoteGameEnd(mSoul.Ctx, &message.ReqVoteGameEnd{Yes: true})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", end)
}

// ActionNewRound 回合开始
func (mSoul *Majsoul) ActionNewRound(action *message.ActionNewRound) {
	log.Printf("%+v", action)
	mSoul.tiles = action.Tiles

	// 如果是庄家
	if len(action.Tiles) == 14 {
		operation, err := mSoul.InputOperation(mSoul.Ctx, &message.ReqSelfOperation{
			Type:    majsoul.Discard, // 打出牌
			Tile:    mSoul.tiles[len(mSoul.tiles)-1],
			Moqie:   true,
			Timeuse: 1,
		})
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Printf("%+v", operation)
	}
}

// ActionDealTile 摸牌
func (mSoul *Majsoul) ActionDealTile(action *message.ActionDealTile) {
	log.Printf("%+v", action)
	tsumoKiri := mSoul.tiles[len(mSoul.tiles)-1] == action.Tile
	_, err := mSoul.InputOperation(mSoul.Ctx, &message.ReqSelfOperation{
		Type:    majsoul.Discard, // 打出牌
		Tile:    action.Tile,
		Moqie:   tsumoKiri,
		Timeuse: 1,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
}

// ActionChiPengGang 吃碰杠
func (mSoul *Majsoul) ActionChiPengGang(action *message.ActionChiPengGang) {
	log.Printf("%+v", action)
	_, err := mSoul.InputOperation(mSoul.Ctx, &message.ReqSelfOperation{
		CancelOperation: true, // 取消操作
		Timeuse:         1,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	mSoul := NewMajsoul()
	resLogin, err := mSoul.Login("", "")
	if err != nil {
		log.Fatal(err)
		return
	}

	// 检查是否在游戏中
	if resLogin.Account.RoomId != 0 {
		log.Println("在游戏中")
		ReConnect(mSoul)
	}

	select {
	case <-mSoul.Ctx.Done():
	}
}

func ReConnect(mSoul *Majsoul) {
	// 重新连接
	_, err := mSoul.ReadyPlay(mSoul.Ctx, &message.ReqRoomReady{Ready: true})
	if err != nil {
		log.Fatal(err)
		return
	}
}
