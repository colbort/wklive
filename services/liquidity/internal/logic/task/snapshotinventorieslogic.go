package tasklogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SnapshotInventoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSnapshotInventoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SnapshotInventoriesLogic {
	return &SnapshotInventoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SnapshotInventoriesLogic) SnapshotInventories(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	if err := validateTask(in); err != nil {
		return nil, err
	}
	return taskDependencyUnavailable("inventory snapshot"), nil
}
