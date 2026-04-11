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

type GetMarginAccountListAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMarginAccountListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginAccountListAdminLogic {
	return &GetMarginAccountListAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMarginAccountListAdminLogic) GetMarginAccountListAdmin(req *types.GetMarginAccountListAdminReq) (resp *types.GetMarginAccountListAdminResp, err error) {
	return logicutil.Proxy[types.GetMarginAccountListAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetMarginAccountListAdmin)
}
