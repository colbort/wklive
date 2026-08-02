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

type GetMarginSnapshotListAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMarginSnapshotListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarginSnapshotListAdminLogic {
	return &GetMarginSnapshotListAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMarginSnapshotListAdminLogic) GetMarginSnapshotListAdmin(req *types.GetMarginSnapshotListAdminReq) (resp *types.GetMarginSnapshotListAdminResp, err error) {
	return logicutil.Proxy[types.GetMarginSnapshotListAdminResp](l.ctx, req, l.svcCtx.TradeCli.GetMarginSnapshotListAdmin)
}
