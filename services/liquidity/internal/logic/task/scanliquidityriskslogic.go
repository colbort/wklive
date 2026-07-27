package tasklogic

import (
	"context"
	"wklive/services/liquidity/internal/logic/helpers"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ScanLiquidityRisksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewScanLiquidityRisksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScanLiquidityRisksLogic {
	return &ScanLiquidityRisksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ScanLiquidityRisksLogic) ScanLiquidityRisks(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	if err := helpers.ValidateTask(in); err != nil {
		return nil, err
	}
	return helpers.TaskDependencyUnavailable("liquidity risk scanner"), nil
}
