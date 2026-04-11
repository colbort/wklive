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

type GetFillListAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFillListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFillListAdminLogic {
	return &GetFillListAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFillListAdminLogic) GetFillListAdmin(req *types.GetFillListAdminReq) (resp *types.GetFillListAdminResp, err error) {
	return logicutil.Proxy[types.GetFillListAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetFillListAdmin)
}
