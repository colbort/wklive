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

type AdminGetMarketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetMarketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetMarketLogic {
	return &AdminGetMarketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetMarketLogic) AdminGetMarket(req *types.GetMarketReq) (resp *types.GetMarketResp, err error) {
	return logicutil.Proxy[types.GetMarketResp](l.ctx, req, l.svcCtx.OptionCli.AdminGetMarket)
}
