// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"fmt"
	"strconv"
	"wklive/common/utils"
	"wklive/liquidity-admin-api/internal/config"
	liquidityclient "wklive/services/liquidity/client/admin"
	systemclient "wklive/services/system/client/admin"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServiceContext struct {
	Config       config.Config
	SystemCli    systemclient.Admin
	LiquidityCli liquidityclient.Admin
}

func NewServiceContext(c config.Config) *ServiceContext {
	options := zrpc.WithUnaryClientInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		pairs := make([]string, 0, 6)
		if id, err := utils.GetUserIdFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyUid, strconv.FormatInt(id, 10))
		}
		if username, err := utils.GetUsernameFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyUsername, username)
		}
		if tenantId, err := utils.GetTenantIdFromCtx(ctx); err == nil {
			pairs = append(pairs, utils.CtxKeyTenantId, fmt.Sprintf("%d", tenantId))
		}
		if len(pairs) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
	return &ServiceContext{
		Config:       c,
		SystemCli:    systemclient.NewAdmin(zrpc.MustNewClient(c.SystemRpc, options)),
		LiquidityCli: liquidityclient.NewAdmin(zrpc.MustNewClient(c.LiquidityRpc, options)),
	}
}
