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

type AdminUpdateMarketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateMarketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateMarketLogic {
	return &AdminUpdateMarketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateMarketLogic) AdminUpdateMarket(req *types.UpdateMarketReq) (resp *types.OptionAdminCommonResp, err error) {
	return logicutil.Proxy[types.OptionAdminCommonResp](l.ctx, req, l.svcCtx.OptionCli.AdminUpdateMarket)
}
