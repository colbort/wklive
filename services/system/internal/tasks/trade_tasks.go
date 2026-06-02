package tasks

import (
	"context"
	"fmt"
	"time"

	"wklive/proto/trade"
	"wklive/services/system/internal/global"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

const tradeTaskRPCTimeout = 5 * time.Minute

func init() {
	cronx.Register("trade.ProcessOrderMatching", "订单撮合", runTradeProcessOrderMatching)
	cronx.Register("trade.ProcessPositions", "仓位处理", runTradeProcessPositions)
	cronx.Register("trade.ProcessContractSettlements", "合约结算处理", runTradeProcessContractSettlements)
	cronx.Register("trade.ProcessTradeEvents", "交易事件处理", runTradeProcessTradeEvents)
	cronx.Register("trade.ExpireRiskLimits", "风控限制过期恢复", runTradeExpireRiskLimits)
}

func runTradeProcessOrderMatching(ctx context.Context, job *models.SysJob) error {
	return callTradeTask(ctx, job, "process order matching", global.TradeTaskCli.ProcessOrderMatching)
}

func runTradeProcessPositions(ctx context.Context, job *models.SysJob) error {
	return callTradeTask(ctx, job, "process positions", global.TradeTaskCli.ProcessPositions)
}

func runTradeProcessContractSettlements(ctx context.Context, job *models.SysJob) error {
	return callTradeTask(ctx, job, "process contract settlements", global.TradeTaskCli.ProcessContractSettlements)
}

func runTradeProcessTradeEvents(ctx context.Context, job *models.SysJob) error {
	return callTradeTask(ctx, job, "process trade events", global.TradeTaskCli.ProcessTradeEvents)
}

func runTradeExpireRiskLimits(ctx context.Context, job *models.SysJob) error {
	return callTradeTask(ctx, job, "expire risk limits", global.TradeTaskCli.ExpireRiskLimits)
}

func callTradeTask(
	ctx context.Context,
	_ *models.SysJob,
	action string,
	fn func(context.Context, *trade.TradeTaskReq, ...grpcCallOption) (*trade.TradeTaskResp, error),
) error {
	rpcCtx, cancel := context.WithTimeout(ctx, tradeTaskRPCTimeout)
	defer cancel()

	result, err := fn(rpcCtx, &trade.TradeTaskReq{TenantId: 0})
	if err != nil {
		return err
	}
	if result == nil || result.Base == nil {
		return fmt.Errorf("trade %s failed, empty response", action)
	}
	if result.Base.Code != 200 {
		return fmt.Errorf("trade %s failed, code: %d, message: %s", action, result.Base.Code, result.Base.Msg)
	}
	return nil
}
