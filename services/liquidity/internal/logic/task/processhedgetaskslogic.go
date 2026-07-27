package tasklogic

import (
	"context"
	"wklive/services/liquidity/internal/logic/helpers"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessHedgeTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessHedgeTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessHedgeTasksLogic {
	return &ProcessHedgeTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProcessHedgeTasksLogic) ProcessHedgeTasks(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	if err := helpers.ValidateTask(in); err != nil {
		return nil, err
	}
	return helpers.TaskDependencyUnavailable("hedge task"), nil
}
