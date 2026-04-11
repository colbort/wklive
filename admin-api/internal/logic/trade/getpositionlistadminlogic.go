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

type GetPositionListAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPositionListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionListAdminLogic {
	return &GetPositionListAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPositionListAdminLogic) GetPositionListAdmin(req *types.GetPositionListAdminReq) (resp *types.GetPositionListAdminResp, err error) {
	return logicutil.Proxy[types.GetPositionListAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetPositionListAdmin)
}
