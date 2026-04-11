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

type GetCancelLogListAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCancelLogListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCancelLogListAdminLogic {
	return &GetCancelLogListAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCancelLogListAdminLogic) GetCancelLogListAdmin(req *types.GetCancelLogListAdminReq) (resp *types.GetCancelLogListAdminResp, err error) {
	return logicutil.Proxy[types.GetCancelLogListAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetCancelLogListAdmin)
}
