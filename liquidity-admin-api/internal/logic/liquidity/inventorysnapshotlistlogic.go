// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InventorySnapshotListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInventorySnapshotListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InventorySnapshotListLogic {
	return &InventorySnapshotListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InventorySnapshotListLogic) InventorySnapshotList(req *types.InventoryQuery) (resp *types.InventoryListResp, err error) {
	return inventorySnapshotList(l.ctx, l.svcCtx, req)
}
