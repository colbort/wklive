// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LatestInventoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLatestInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LatestInventoryLogic {
	return &LatestInventoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LatestInventoryLogic) LatestInventory(req *types.LatestInventoryReq) (resp *types.InventoryResp, err error) {
	return latestInventory(l.ctx, l.svcCtx, req)
}
