package tasks

import (
	"context"

	"wklive/common/tasks"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("liquidity.RefreshQuotes", "做市报价刷新", runLiquidityRefreshQuotes)
	cronx.Register("liquidity.RecoverQuoteOrders", "做市报价状态恢复", runLiquidityRecoverQuoteOrders)
}

func runLiquidityRefreshQuotes(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceLiquidity, tasks.ActionLiquidityRefreshQuotes)
}

func runLiquidityRecoverQuoteOrders(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceLiquidity, tasks.ActionLiquidityRecoverQuoteOrders)
}
