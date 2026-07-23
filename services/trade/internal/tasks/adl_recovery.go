package tasks

import (
	"context"
	"time"

	logic "wklive/services/trade/internal/logic/task"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartADLRecovery continuously resumes both ADL child executions and their
// parent liquidation sagas after crashes, timeouts, or downstream failures.
func StartADLRecovery(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := logic.NewProcessLiquidationsLogic(ctx, svcCtx).RecoverADLExecutions(100); err != nil {
					logx.Errorf("ADL recovery scan failed: %v", err)
				}
				if err := logic.NewProcessLiquidationsLogic(ctx, svcCtx).RecoverLiquidations(100); err != nil {
					logx.Errorf("liquidation saga recovery scan failed: %v", err)
				}
			}
		}
	}()
}
