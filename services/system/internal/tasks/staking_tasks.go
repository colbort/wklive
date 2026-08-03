package tasks

import (
	"context"

	"wklive/common/tasks"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("staking.ProcessRewardsAndSettleOrders", "质押收益发放/到期结算", runStakingProcessRewardsAndSettleOrders)
	cronx.Register("staking.ReconcileStaking", "质押每日账实对账", runStakingReconcile)
}

func runStakingReconcile(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceStaking, tasks.ActionStakingReconcile)
}

func runStakingProcessRewardsAndSettleOrders(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceStaking, tasks.ActionStakingProcessRewardsAndSettleOrders)
}
