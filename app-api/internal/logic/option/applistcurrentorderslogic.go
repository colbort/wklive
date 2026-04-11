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

type AppListCurrentOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListCurrentOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListCurrentOrdersLogic {
	return &AppListCurrentOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListCurrentOrdersLogic) AppListCurrentOrders(req *types.AppListCurrentOrdersReq) (resp *types.AppListCurrentOrdersResp, err error) {
	return logicutil.Proxy[types.AppListCurrentOrdersResp](l.ctx, req, l.svcCtx.OptionCli.AppListCurrentOrders)
}
