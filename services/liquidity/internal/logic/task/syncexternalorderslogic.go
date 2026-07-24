package tasklogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncExternalOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncExternalOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncExternalOrdersLogic {
	return &SyncExternalOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SyncExternalOrdersLogic) SyncExternalOrders(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	if err := validateTask(in); err != nil {
		return nil, err
	}
	return taskDependencyUnavailable("external order synchronization"), nil
}
