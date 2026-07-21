package tasks

import (
	"context"
	"time"

	"wklive/services/trade/internal/logic"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// StartADLRecovery continuously resumes ADL sagas left between the asset and
// position steps by a crash, timeout, or temporary downstream failure.
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
			}
		}
	}()
}
