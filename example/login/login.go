package main

import (
	"context"
	"fmt"
	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/message"
	"os"
	"time"
)

func main() {
	account, exists := os.LookupEnv("account")
	if !exists {
		panic("account is required.")
	}

	password, exists := os.LookupEnv("password")
	if !exists {
		panic("password is required.")
	}

	majSoul := majsoul.NewMajSoul(&majsoul.Config{ProxyAddress: ""})
	{ // 寻找可用服务器
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err := majSoul.LookupGateway(ctx, majsoul.ServerAddressList)
		if err != nil {
			panic(err)
		}
	}

	{ // 登录
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		resLogin, err := majSoul.Login(ctx, account, password)
		if err != nil {
			panic(err)
		}
		if resLogin.Error != nil {
			panic(resLogin.Error)
		}
	}

	{ // 获取好友列表
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		friendList, err := majSoul.LobbyClient.FetchFriendList(ctx, &message.ReqCommon{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v", friendList)
	}
}
