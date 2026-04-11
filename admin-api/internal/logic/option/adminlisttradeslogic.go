// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListTradesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListTradesLogic {
	return &AdminListTradesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListTradesLogic) AdminListTrades(req *types.ListTradesReq) (resp *types.ListTradesResp, err error) {
	return logicutil.Proxy[types.ListTradesResp](l.ctx, req, l.svcCtx.OptionCli.AdminListTrades)
}
