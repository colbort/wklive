package adminlogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInventorySnapshotListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInventorySnapshotListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInventorySnapshotListLogic {
	return &GetInventorySnapshotListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetInventorySnapshotListLogic) GetInventorySnapshotList(in *liquidity.GetInventorySnapshotListReq) (*liquidity.GetInventorySnapshotListResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.GetInventorySnapshotListResp{}, nil
}
