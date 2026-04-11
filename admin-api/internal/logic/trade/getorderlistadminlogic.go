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

type GetOrderListAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderListAdminLogic {
	return &GetOrderListAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderListAdminLogic) GetOrderListAdmin(req *types.GetOrderListAdminReq) (resp *types.GetOrderListAdminResp, err error) {
	return logicutil.Proxy[types.GetOrderListAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetOrderListAdmin)
}
