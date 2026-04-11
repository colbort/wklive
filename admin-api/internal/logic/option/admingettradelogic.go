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

type AdminGetTradeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetTradeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetTradeLogic {
	return &AdminGetTradeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetTradeLogic) AdminGetTrade(req *types.GetTradeReq) (resp *types.GetTradeResp, err error) {
	return logicutil.Proxy[types.GetTradeResp](l.ctx, req, l.svcCtx.OptionCli.AdminGetTrade)
}
