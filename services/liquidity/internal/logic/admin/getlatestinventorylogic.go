package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestInventoryLogic {
	return &GetLatestInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLatestInventoryLogic) GetLatestInventory(in *liquidity.GetLatestInventoryReq) (*liquidity.InventorySnapshotResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.InventorySnapshotResp{}, nil
}
