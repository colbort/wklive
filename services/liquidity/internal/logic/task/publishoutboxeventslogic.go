package tasklogic

import (
	"context"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishOutboxEventsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishOutboxEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishOutboxEventsLogic {
	return &PublishOutboxEventsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PublishOutboxEventsLogic) PublishOutboxEvents(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	if err := validateTask(in); err != nil {
		return nil, err
	}
	return taskDependencyUnavailable("outbox publisher"), nil
}
