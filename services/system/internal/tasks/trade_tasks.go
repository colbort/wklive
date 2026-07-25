package tasks

import (
	"context"

	"wklive/common/tasks"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("trade.ProcessOrderMatching", "订单撮合", runTradeProcessOrderMatching)
	cronx.Register("trade.ProcessPositions", "仓位处理", runTradeProcessPositions)
	cronx.Register("trade.ProcessContractSettlements", "合约结算处理", runTradeProcessContractSettlements)
	cronx.Register("trade.ProcessSecondsSettlements", "秒合约结算处理", runTradeProcessSecondsSettlements)
	cronx.Register("trade.ProcessTradeEvents", "交易事件处理", runTradeProcessTradeEvents)
	cronx.Register("trade.ExpireRiskLimits", "风控限制过期恢复", runTradeExpireRiskLimits)
	cronx.Register("trade.ArchiveLiquidityOrders", "归档零成交做市撤单", runTradeArchiveLiquidityOrders)
}

func runTradeProcessOrderMatching(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceTrade, tasks.ActionTradeProcessOrderMatching)
}

func runTradeProcessPositions(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceTrade, tasks.ActionTradeProcessPositions)
}

func runTradeProcessContractSettlements(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceTrade, tasks.ActionTradeProcessContractSettlements)
}

func runTradeProcessSecondsSettlements(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceTrade, tasks.ActionTradeProcessSecondsSettlements)
}

func runTradeProcessTradeEvents(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceTrade, tasks.ActionTradeProcessTradeEvents)
}

func runTradeExpireRiskLimits(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceTrade, tasks.ActionTradeExpireRiskLimits)
}

func runTradeArchiveLiquidityOrders(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceTrade, tasks.ActionTradeArchiveLiquidityOrders)
}
