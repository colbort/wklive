package tasklogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SettleExternalFillsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSettleExternalFillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SettleExternalFillsLogic {
	return &SettleExternalFillsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SettleExternalFillsLogic) SettleExternalFills(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	// todo: add your logic here and delete this line

	return &liquidity.LiquidityTaskResp{}, nil
}
