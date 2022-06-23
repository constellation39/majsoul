# majsoul

## [majsoul](https://game.maj-soul.com/1) 的客户端通信协议Go实现

使用grpc生成了向majsoul服务器请求的通信协议，但是对于majsoul服务器的消息下发使用了更加原始的处理方式。

> current liqi.proto version v0.10.103.w

### 安装

```
go get -u github.com/constellation39/majsoul
```

## 示例

### 向服务器发送消息
```go
// 读取配置文件，该文件声明了 majsoul 应该连接的服务器地址
config, err := majsoul.LoadConfig("majsoul.json")
if err != nil {
	log.Fatal(err)
}

mSoul := majsoul.New(config)

resLogin, err := mSoul.Login("account", "password")
if err != nil {
	log.Fatal(err)
}
if resLogin.Error != nil {
	log.Fatal(resLogin.Error)
}

log.Printf("登录成功")

friendList, err := mSoul.FetchFriendList(mSoul.Ctx, &message.ReqCommon{})
if err != nil {
    log.Fatal(err)
}

log.Printf("好友列表: %+v", friendList)
```

### 获取服务器下发消息
```go

// 首先继承该包的majsoul.Majsoul，然后实现需要监听的方法
type ImplementMajsoul struct {
	*majsoul.Majsoul
}

// NewImplementMajsoul 创建一个 ImplementMajsoul 结构
func NewImplementMajsoul() *ImplementMajsoul {
	config, err := majsoul.LoadConfig("majsoul.json")
	if err != nil {
		log.Fatal(err)
	}
	mSoul := &Majsoul{Majsoul: majsoul.New(config)}
	mSoul.Implement = mSoul // 使用多态实现，如果调用时没有提供外部实现则调用内部的实现，如果没有给 Implement 赋值，则只会调用内部实现
	return mSoul
}

// 以下是在游戏大厅内消息，对应函数在 majsoul/inotify.go 中由 IFNotify 接口定义

// NotifyClientMessage 实现对应客户端消息
// message.NotifyClientMessage filed Type == 1 时为受到邀请
// note: 这个函数的只实现了接受到邀请的通知
func (mSoul *ImplementMajsoul) NotifyClientMessage(notify *message.NotifyClientMessage) {
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
    // 同意
	_, err := mSoul.VoteGameEnd(mSoul.Ctx, &message.ReqVoteGameEnd{Yes: true})
	if err != nil {
		log.Fatal(err)
	}
}

// 以下是游戏桌面消息，对应函数在 majsoul/iaction.go 中由 IFAction 接口定义

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

```

完整的示例文件在 [example](https://github.com/constellation39/majsoul/tree/master/example) 文件中