package tasklogic

import (
	"context"
	"fmt"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessSecondsSettlementsTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessSecondsSettlementsTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessSecondsSettlementsTaskLogic {
	return &ProcessSecondsSettlementsTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ProcessSecondsSettlements runs independently at second-level frequency.
func (l *ProcessSecondsSettlementsTaskLogic) ProcessSecondsSettlements(in *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_seconds_settlements", func(taskCtx context.Context) (*trade.TradeTaskResp, error) {
		l.ctx = taskCtx
		if err := NewProcessSecondsSettlementsLogic(l.ctx, l.svcCtx).Process(in.GetTenantId()); err != nil {
			return nil, fmt.Errorf("seconds settlements: %w", err)
		}
		return helpers.OkTaskResp(), nil
	})
}
