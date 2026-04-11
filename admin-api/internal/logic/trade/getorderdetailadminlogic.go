// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderDetailAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailAdminLogic {
	return &GetOrderDetailAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderDetailAdminLogic) GetOrderDetailAdmin(req *types.GetOrderDetailAdminReq) (resp *types.GetOrderDetailAdminResp, err error) {
	return logicutil.Proxy[types.GetOrderDetailAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetOrderDetailAdmin)
}
