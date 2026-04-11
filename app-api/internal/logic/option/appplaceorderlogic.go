// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppPlaceOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppPlaceOrderLogic {
	return &AppPlaceOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppPlaceOrderLogic) AppPlaceOrder(req *types.AppPlaceOrderReq) (resp *types.AppPlaceOrderResp, err error) {
	return logicutil.Proxy[types.AppPlaceOrderResp](l.ctx, req, l.svcCtx.OptionCli.AppPlaceOrder)
}
