package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/message"
)

// Majsoul 组合库中的 Majsoul 结构
type Majsoul struct {
	*majsoul.Majsoul
	seat  uint32
	tiles []string
}

// NewMajsoul 创建一个 Majsoul 结构
func NewMajsoul() *Majsoul {
	mj, err := majsoul.New(&majsoul.Config{
		ServerAddressList: nil,
		ServerProxy:       "",
		GatewayProxy:      "",
		GameProxy:         "",
		Reconnect:         true,
		ReconnectInterval: time.Second,
		ReconnectNumber:   3,
	})
	if err != nil {
		panic(err)
	}
	mSoul := &Majsoul{Majsoul: mj}
	mSoul.Implement = mSoul // 使用多态实现，如果调用时没有提供外部实现则调用内部的实现，如果没有给 Implement 赋值，则只会调用内部实现
	return mSoul
}

func main() {
	mSoul := NewMajsoul()
	resLogin, err := mSoul.Login("1601198895@qq.com", "miku39..")
	if err != nil {
		log.Fatal(err)
	}
	if resLogin.Error != nil {
		log.Fatal(resLogin.Error)
	}

	// 检查是否在游戏中
	if resLogin.Account.RoomId != 0 {
		log.Println("在游戏中")
	}

	friendList, err := mSoul.FetchFriendList(mSoul.Ctx, &message.ReqCommon{})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("好友列表: %+v", friendList)

}

func (mSoul *Majsoul) NotifyRoomGameStart(notify *message.NotifyRoomGameStart) {
	log.Printf("NotifyRoomGameStart %+v", notify)
	mSoul.Majsoul.NotifyRoomGameStart(notify)

	// 记录自己的座位号
	for i, uid := range mSoul.GameInfo.SeatList {
		if uid == mSoul.Account.AccountId {
			mSoul.seat = uint32(i)
			break
		}
	}
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

	// 加入房间
	_, err = mSoul.JoinRoom(mSoul.Ctx, &message.ReqJoinRoom{
		RoomId:              invitationRoom.RoomID,
		ClientVersionString: mSoul.Version.Web(),
	})
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	// 准备
	_, err = mSoul.ReadyPlay(mSoul.Ctx, &message.ReqRoomReady{Ready: true})
	if err != nil {
		log.Printf("%+v", err)
		return
	}
}

// NotifyEndGameVote 有人发起投降
func (mSoul *Majsoul) NotifyEndGameVote(notify *message.NotifyEndGameVote) {
	_, err := mSoul.VoteGameEnd(mSoul.Ctx, &message.ReqVoteGameEnd{Yes: true})
	if err != nil {
		log.Fatal(err)
	}
}

// ActionNewRound 回合开始
func (mSoul *Majsoul) ActionNewRound(action *message.ActionNewRound) {
	log.Printf("ActionNewRound %+v", action)

	mSoul.tiles = action.Tiles

	// 如果是庄家
	if len(action.Tiles) == 14 {
		time.Sleep(time.Second * 3)
		_, err := mSoul.InputOperation(mSoul.Ctx, &message.ReqSelfOperation{
			Type:    majsoul.Discard, // 打出牌
			Tile:    mSoul.tiles[len(mSoul.tiles)-1],
			Moqie:   true,
			Timeuse: 1,
		})
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}

// ActionDealTile 摸牌
func (mSoul *Majsoul) ActionDealTile(action *message.ActionDealTile) {
	log.Printf("ActionDealTile %+v", action)
	// 如果不是自己摸牌
	if action.Seat != mSoul.seat {
		return
	}
	_, err := mSoul.InputOperation(mSoul.Ctx, &message.ReqSelfOperation{
		Type:      majsoul.Discard, // 打出牌
		Tile:      action.Tile,
		Moqie:     true,
		Timeuse:   1,
		TileState: 0,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
}

// ActionDiscardTile 打牌
func (mSoul *Majsoul) ActionDiscardTile(action *message.ActionDiscardTile) {
	log.Printf("ActionDiscardTile %+v", action)
	if action.Operation != nil {
		return
	}
	_, err := mSoul.InputOperation(mSoul.Ctx, &message.ReqSelfOperation{
		CancelOperation: true, // 取消操作
		Timeuse:         1,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// ActionChiPengGang 吃碰杠
func (mSoul *Majsoul) ActionChiPengGang(action *message.ActionChiPengGang) {
	log.Printf("ActionChiPengGang %+v", action)
}
