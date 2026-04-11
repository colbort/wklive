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

type AppListTradesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListTradesLogic {
	return &AppListTradesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppListTradesLogic) AppListTrades(req *types.AppListTradesReq) (resp *types.AppListTradesResp, err error) {
	return logicutil.Proxy[types.AppListTradesResp](l.ctx, req, l.svcCtx.OptionCli.AppListTrades)
}
