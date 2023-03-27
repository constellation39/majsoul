package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/constellation39/majsoul"
	"github.com/constellation39/majsoul/logger"
	"github.com/constellation39/majsoul/message"
	"go.uber.org/zap"
)

type Majsoul struct {
	*majsoul.Majsoul
}

var (
	account  = flag.String("account", "", "majsoul login when the account.")
	password = flag.String("password", "", "majsoul login when the password.")
)

func main() {
	flag.Parse()
	logger.EnableDevelopment()

	if *account == "" {
		logger.Error("account is required.")
		os.Exit(1)
	}

	if *password == "" {
		logger.Error("password is required.")
		os.Exit(1)
	}

	// 初始化一个客户端
	ctx := context.Background()
	subClient, err := majsoul.New(ctx)
	if err != nil {
		logger.Error("majsoul client is not created.", zap.Error(err))
		os.Exit(1)
	}
	client := &Majsoul{Majsoul: subClient}
	// 使用了多态的方式实现
	// 需要监听雀魂服务器下发通知时，需要实现这个接口 majsoul.Implement
	// majsoul.Majsoul 原生实现了这个接口，只需要重写需要的方法即可
	subClient.Implement = client
	logger.Info("majsoul client is created.", zap.Reflect("ServerAddress", subClient.ServerAddress))

	// 按照雀魂web端的请求进行模拟
	timeOutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	resLogin, err := client.Login(timeOutCtx, *account, *password)
	if err != nil {
		logger.Error("majsoul login error.", zap.Error(err))
		return
	}
	logger.Info("majsoul login.", zap.Reflect("resLogin", resLogin))

	resFetchLastPrivacy, err := client.FetchLastPrivacy(ctx, &message.ReqFetchLastPrivacy{})
	if err != nil {
		logger.Error("majsoul FetchLastPrivacy error.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchLastPrivacy.", zap.Reflect("resFetchLastPrivacy", resFetchLastPrivacy))

	resFetchServerTime, err := client.FetchServerTime(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchServerTime error.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchServerTime.", zap.Reflect("resFetchServerTime", resFetchServerTime))

	resServerSettings, err := client.FetchServerSettings(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchServerSettings error.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchServerSettings.", zap.Reflect("resServerSettings", resServerSettings))

	resConnectionInfo, err := client.FetchConnectionInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchConnectionInfo error.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchConnectionInfo.", zap.Reflect("resConnectionInfo", resConnectionInfo))

	resClientValue, err := client.FetchClientValue(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchClientValue error.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchClientValue.", zap.Reflect("resClientValue", resClientValue))

	resFriendList, err := client.FetchFriendList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchFriendList error.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchFriendList.", zap.Reflect("resFriendList", resFriendList))

	resFriendApplyList, err := client.FetchFriendApplyList(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchFriendApplyList error.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchFriendApplyList.", zap.Reflect("resFriendApplyList", resFriendApplyList))

	resFetchrecentFriend, err := client.FetchRecentFriend(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchRecentFriend.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchRecentFriend.", zap.Reflect("resFetchrecentFriend", resFetchrecentFriend))

	resMailInfo, err := client.FetchMailInfo(ctx, &message.ReqCommon{})
	if err != nil {
		logger.Error("majsoul FetchMailInfo error.", zap.Error(err))
		return
	}
	logger.Info("majsoul FetchMailInfo.", zap.Reflect("resMailInfo", resMailInfo))
}
