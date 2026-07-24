package tasklogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessReconciliationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessReconciliationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessReconciliationsLogic {
	return &ProcessReconciliationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProcessReconciliationsLogic) ProcessReconciliations(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.LiquidityTaskResp{}, nil
}
