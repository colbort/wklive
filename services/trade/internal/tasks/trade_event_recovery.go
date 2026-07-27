package tasks

import (
	"context"
	"time"

	"wklive/common/i18n"
	"wklive/proto/trade"
	logic "wklive/services/trade/internal/logic/task"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const tradeEventRecoveryInterval = 5 * time.Second

// StartTradeEventRecovery provides an in-process recovery safety net for
// settlement, reservation-release and terminating-order state machines.
// System cron/Kafka remains the normal trigger; the shared task lock makes
// concurrent triggers harmless. Running once immediately is important after a
// restart because old CANCELING orders otherwise wait indefinitely when the
// scheduler job is absent or task delivery is interrupted.
func StartTradeEventRecovery(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		run := func() {
			resp, err := logic.NewProcessTradeEventsLogic(ctx, svcCtx).
				ProcessTradeEvents(&trade.TradeTaskReq{})
			if err != nil {
				logx.Errorf("trade event recovery scan failed: %v", err)
				return
			}
			if resp == nil || resp.GetBase() == nil {
				logx.Errorf("trade event recovery scan returned an empty response")
				return
			}
			if resp.GetBase().GetCode() == int32(i18n.SyncTaskAlreadyRunning) {
				return
			}
			if resp.GetBase().GetCode() != 200 {
				logx.Errorf(
					"trade event recovery scan failed, code=%d msg=%s",
					resp.GetBase().GetCode(),
					resp.GetBase().GetMsg(),
				)
			}
		}

		run()
		ticker := time.NewTicker(tradeEventRecoveryInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
