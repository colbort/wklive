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

type GetPositionHistoryListAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPositionHistoryListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionHistoryListAdminLogic {
	return &GetPositionHistoryListAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPositionHistoryListAdminLogic) GetPositionHistoryListAdmin(req *types.GetPositionHistoryListAdminReq) (resp *types.GetPositionHistoryListAdminResp, err error) {
	return logicutil.Proxy[types.GetPositionHistoryListAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetPositionHistoryListAdmin)
}
