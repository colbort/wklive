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

type AdminGetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetOrderLogic {
	return &AdminGetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetOrderLogic) AdminGetOrder(req *types.GetOrderReq) (resp *types.GetOrderResp, err error) {
	return logicutil.Proxy[types.GetOrderResp](l.ctx, req, l.svcCtx.OptionCli.AdminGetOrder)
}
