package tasklogic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/common/worklease"
	"wklive/services/liquidity/internal/logic/helpers"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishOutboxEventsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	owner string
}

func NewPublishOutboxEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishOutboxEventsLogic {
	return &PublishOutboxEventsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		owner:  worklease.NewOwner("liquidity-outbox"),
	}
}

func (l *PublishOutboxEventsLogic) PublishOutboxEvents(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	if err := helpers.ValidateTask(in); err != nil {
		return nil, err
	}
	batchSize := int64(in.BatchSize)
	if batchSize <= 0 || batchSize > 500 {
		batchSize = 100
	}
	now := time.Now()
	rows, err := l.svcCtx.EventOutboxModel.FindRunnable(
		l.ctx, now.UnixMilli(), now.Add(-time.Minute).UnixMilli(), batchSize,
	)
	if err != nil {
		return nil, err
	}
	resp := &liquidity.LiquidityTaskResp{
		Base: helper.OkResp(), ScannedCount: int64(len(rows)),
	}
	for _, row := range rows {
		claimTime := time.Now()
		claimed, claimErr := l.svcCtx.EventOutboxModel.Claim(
			l.ctx, row.Id, l.owner, claimTime.UnixMilli(), claimTime.Add(-time.Minute).UnixMilli(),
		)
		if claimErr != nil {
			return nil, claimErr
		}
		if !claimed {
			continue
		}
		publishErr := l.svcCtx.OutboxPublisher.PublishKey(
			l.ctx, row.Topic, []byte(row.MessageKey), row.Payload,
		)
		if publishErr != nil {
			updated, markErr := l.svcCtx.EventOutboxModel.MarkFailed(
				l.ctx, row, l.owner, publishErr.Error(), time.Now().UnixMilli(),
			)
			if markErr != nil {
				return nil, markErr
			}
			if !updated {
				return nil, fmt.Errorf("liquidity outbox lease lost while marking failure: %s", row.EventNo)
			}
			resp.FailedCount++
			continue
		}
		updated, markErr := l.svcCtx.EventOutboxModel.MarkSuccess(
			l.ctx, row.Id, l.owner, time.Now().UnixMilli(),
		)
		if markErr != nil {
			return nil, markErr
		}
		if !updated {
			return nil, fmt.Errorf("liquidity outbox lease lost while marking success: %s", row.EventNo)
		}
		resp.SuccessCount++
	}
	return resp, nil
}
