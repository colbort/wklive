// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"wklive/admin-api/internal/config"
	"wklive/admin-api/internal/ws"
	mq "wklive/common/mq/kafka"
	"wklive/common/reqenc"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/chat"
	"wklive/proto/market"
	"wklive/proto/option"
	"wklive/proto/payment"
	"wklive/proto/staking"
	"wklive/proto/system"
	"wklive/proto/trade"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServiceContext struct {
	Config            config.Config
	SystemCli         system.AdminClient
	ChatCli           chat.PlatformClient
	UserCli           user.AdminClient
	PaymentCli        payment.AdminClient
	MarketCli         market.AdminClient
	AssetCli          asset.AdminClient
	OptionCli         option.AdminClient
	StakingCli        staking.AdminClient
	TradeCli          trade.AdminClient
	NotificationHub   *ws.Hub
	NotificationStore *ws.NotificationStore
	RequestEncryption *reqenc.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	options := zrpc.WithUnaryClientInterceptor(func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		pairs := []string{utils.CtxKeySubjectDomain, utils.SubjectDomainSystem}
		if userId, err := utils.GetUserIdFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyUid, strconv.FormatInt(userId, 10))
		}
		if username, err := utils.GetUsernameFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyUsername, username)
		}
		if tenantId, err := utils.GetTenantIdFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyTenantId, fmt.Sprintf("%d", tenantId))
		}
		if userType, err := utils.GetUserTypeFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyUserType, strconv.FormatInt(userType, 10))
		}
		if tenantCode, err := utils.GetUsernameFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyTenantCode, tenantCode)
		}
		if len(pairs) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
	systemCli := zrpc.MustNewClient(c.SystemRpc, options)
	chatCli := zrpc.MustNewClient(c.ChatRpc, options)
	userCli := zrpc.MustNewClient(c.UserRpc, options)
	paymentCli := zrpc.MustNewClient(c.PaymentRpc, options)
	marketCli := zrpc.MustNewClient(c.MarketRpc, options)
	assetCli := zrpc.MustNewClient(c.AssetRpc, options)
	optionCli := zrpc.MustNewClient(c.OptionRpc, options)
	stakingCli := zrpc.MustNewClient(c.StakingRpc, options)
	tradeCli := zrpc.MustNewClient(c.TradeRpc, options)
	systemAdminClient := system.NewAdminClient(systemCli.Conn())
	notificationHub := ws.NewHub()
	go notificationHub.Run()
	var notificationStore *ws.NotificationStore
	if len(c.MQ.Brokers) > 0 {
		mqConfig := mq.ForService(c.MQ, c.Name)
		notificationStore = ws.NewNotificationStore(
			systemAdminClient,
			mq.MustNewPublisher(mqConfig),
			ws.NotificationStoreConfig{
				AckTimeout:       time.Duration(c.NotificationWS.AckTimeoutMinutes) * time.Minute,
				EscalationRepeat: time.Duration(c.NotificationWS.EscalationIntervalMinutes) * time.Minute,
				EscalationMax:    c.NotificationWS.EscalationMaxLevel,
				EscalationPoll:   time.Duration(c.NotificationWS.EscalationPollSeconds) * time.Second,
			},
		)
		groupID := "admin-notifications-" + hostname()
		subscriber := mq.MustNewBroadcastSubscriber(mqConfig, groupID)
		go ws.SubscribeMQ(context.Background(), subscriber, notificationHub, notificationStore)
		go notificationStore.RunEscalation(context.Background())
	} else {
		logx.Info("admin notification mq is not configured, skip subscription")
	}

	return &ServiceContext{
		Config:            c,
		SystemCli:         systemAdminClient,
		ChatCli:           chat.NewPlatformClient(chatCli.Conn()),
		UserCli:           user.NewAdminClient(userCli.Conn()),
		PaymentCli:        payment.NewAdminClient(paymentCli.Conn()),
		MarketCli:         market.NewAdminClient(marketCli.Conn()),
		AssetCli:          asset.NewAdminClient(assetCli.Conn()),
		OptionCli:         option.NewAdminClient(optionCli.Conn()),
		StakingCli:        staking.NewAdminClient(stakingCli.Conn()),
		TradeCli:          trade.NewAdminClient(tradeCli.Conn()),
		NotificationHub:   notificationHub,
		NotificationStore: notificationStore,
	}
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "unknown"
	}
	return name
}
