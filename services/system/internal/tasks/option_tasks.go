package tasks

import (
	"context"

	"wklive/common/tasks"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("option.ProcessContractLifecycle", "期权合约生命周期处理", runOptionProcessContractLifecycle)
	cronx.Register("option.CleanMarketSnapshots", "期权行情快照清理", runOptionCleanMarketSnapshots)
}

func runOptionProcessContractLifecycle(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionProcessContractLifecycle)
}

func runOptionCleanMarketSnapshots(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceOption, tasks.ActionOptionCleanMarketSnapshots)
}
