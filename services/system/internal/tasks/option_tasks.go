package tasks

import (
	"context"

	"wklive/common/tasks"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("option.ProcessAssetInstructions", "期权资产指令执行与重试", runOptionProcessAssetInstructions)
	cronx.Register("option.ProcessTradeEvents", "期权成交持仓事件处理", runOptionProcessTradeEvents)
	cronx.Register("option.ProcessRiskAccounts", "期权卖方风险账户扫描", runOptionProcessRiskAccounts)
	cronx.Register("option.ProcessLiquidations", "期权卖方强平处理", runOptionProcessLiquidations)
	cronx.Register("option.ProcessExercises", "期权主动行权清算", runOptionProcessExercises)
	cronx.Register("option.ProcessContractLifecycle", "期权合约生命周期处理", runOptionProcessContractLifecycle)
	cronx.Register("option.ProcessCorporateActions", "期权公司行动持仓迁移", runOptionProcessCorporateActions)
	cronx.Register("option.ProcessDailyReconciliation", "期权日终钱包镜像对账", runOptionProcessDailyReconciliation)
	cronx.Register("option.CleanMarketSnapshots", "期权行情快照清理", runOptionCleanMarketSnapshots)
}

func runOptionProcessAssetInstructions(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessAssetInstructions)
}

func runOptionProcessTradeEvents(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessTradeEvents)
}

func runOptionProcessRiskAccounts(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessRiskAccounts)
}

func runOptionProcessLiquidations(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessLiquidations)
}

func runOptionProcessExercises(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessExercises)
}

func runOptionProcessContractLifecycle(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessContractLifecycle)
}

func runOptionProcessCorporateActions(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessCorporateActions)
}

func runOptionProcessDailyReconciliation(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessDailyReconciliation)
}

func runOptionCleanMarketSnapshots(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionCleanMarketSnapshots)
}
