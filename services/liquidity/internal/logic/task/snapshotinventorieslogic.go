package tasklogic

import (
	"context"
	"wklive/services/liquidity/internal/logic/helpers"

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
	if err := helpers.ValidateTask(in); err != nil {
		return nil, err
	}
	return helpers.TaskDependencyUnavailable("inventory snapshot"), nil
}
