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

type GetPositionDetailAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPositionDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionDetailAdminLogic {
	return &GetPositionDetailAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPositionDetailAdminLogic) GetPositionDetailAdmin(req *types.GetPositionDetailAdminReq) (resp *types.GetPositionDetailAdminResp, err error) {
	return logicutil.Proxy[types.GetPositionDetailAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetPositionDetailAdmin)
}
