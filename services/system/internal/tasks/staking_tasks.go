package tasks

import (
	"context"

	"wklive/common/tasks"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("staking.ProcessRewardsAndSettleOrders", "质押收益发放/到期结算", runStakingProcessRewardsAndSettleOrders)
}

func runStakingProcessRewardsAndSettleOrders(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceStaking, tasks.ActionStakingProcessRewardsAndSettleOrders)
}
