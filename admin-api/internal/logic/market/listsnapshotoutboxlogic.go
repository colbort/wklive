// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package market

import (
	"context"
	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSnapshotOutboxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSnapshotOutboxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSnapshotOutboxLogic {
	return &ListSnapshotOutboxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSnapshotOutboxLogic) ListSnapshotOutbox(req *types.ListSnapshotOutboxReq) (resp *types.ListSnapshotOutboxResp, err error) {
	return logicutil.Proxy[types.ListSnapshotOutboxResp](l.ctx, req, l.svcCtx.MarketCli.ListSnapshotOutbox)
}
