package tasklogic

import (
	"context"
	"time"
	"wklive/services/liquidity/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshQuotesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshQuotesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshQuotesLogic {
	return &RefreshQuotesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshQuotesLogic) RefreshQuotes(in *liquidity.LiquidityTaskReq) (*liquidity.LiquidityTaskResp, error) {
	if err := helpers.ValidateTask(in); err != nil {
		return nil, err
	}
	created, prepareFailed, err := prepareInternalQuoteCycles(l.ctx, l.svcCtx, in)
	if err != nil {
		return nil, err
	}
	resp, err := processInternalQuotes(l.ctx, l.svcCtx, in, false)
	if err != nil {
		return nil, err
	}
	if resp.Base == nil {
		resp.Base = helper.OkResp()
	}
	resp.ScannedCount += created
	resp.FailedCount += prepareFailed
	if err := l.svcCtx.QuoteCycleModel.RefreshExecutionResults(l.ctx, in.ConfigId, time.Now().UnixMilli()); err != nil {
		return nil, err
	}
	return resp, nil
}
