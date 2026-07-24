package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
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
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	row, err := l.svcCtx.InventorySnapshotModel.FindLatest(
		l.ctx, in.TenantId, in.ConfigId, in.ProviderId, int64(in.Source),
	)
	if err != nil {
		return nil, err
	}
	return &liquidity.InventorySnapshotResp{Base: helper.OkResp(), Data: helpers.InventoryToProto(row)}, nil
}
