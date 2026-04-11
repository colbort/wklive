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

type GetFillDetailAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFillDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFillDetailAdminLogic {
	return &GetFillDetailAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFillDetailAdminLogic) GetFillDetailAdmin(req *types.GetFillDetailAdminReq) (resp *types.GetFillDetailAdminResp, err error) {
	return logicutil.Proxy[types.GetFillDetailAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetFillDetailAdmin)
}
