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

type AppCancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCancelOrderLogic {
	return &AppCancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppCancelOrderLogic) AppCancelOrder(req *types.AppCancelOrderReq) (resp *types.AppCommonResp, err error) {
	return logicutil.Proxy[types.AppCommonResp](l.ctx, req, l.svcCtx.OptionCli.AppCancelOrder)
}
