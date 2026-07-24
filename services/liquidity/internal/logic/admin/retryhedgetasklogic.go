package adminlogic

import (
	"context"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryHedgeTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryHedgeTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryHedgeTaskLogic {
	return &RetryHedgeTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RetryHedgeTaskLogic) RetryHedgeTask(in *liquidity.RetryHedgeTaskReq) (*liquidity.CommonResp, error) {
	row, err := l.svcCtx.HedgeTaskModel.FindOne(l.ctx, in.HedgeTaskId)
	if err != nil {
		return nil, err
	}
	if row.TenantId != in.TenantId {
		return nil, fmt.Errorf("hedge task not found")
	}
	if row.Version != in.Version {
		return nil, fmt.Errorf("hedge task version conflict")
	}
	if row.Status != int64(liquidity.HedgeStatus_HEDGE_STATUS_FAILED) {
		return nil, fmt.Errorf("only failed hedge task can be retried")
	}
	row.Status, row.NextRetryAt, row.LastErrorMsg = int64(liquidity.HedgeStatus_HEDGE_STATUS_PENDING), 0, ""
	row.Version++
	row.UpdateTimes = time.Now().UnixMilli()
	if err := l.svcCtx.HedgeTaskModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}
