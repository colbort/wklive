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

type AppListHistoryOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListHistoryOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListHistoryOrdersLogic {
	return &AppListHistoryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListHistoryOrdersLogic) AppListHistoryOrders(req *types.AppListHistoryOrdersReq) (resp *types.AppListHistoryOrdersResp, err error) {
	return logicutil.Proxy[types.AppListHistoryOrdersResp](l.ctx, req, l.svcCtx.OptionCli.AppListHistoryOrders)
}
