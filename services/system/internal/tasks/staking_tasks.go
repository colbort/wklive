package tasks

import (
	"context"
	"fmt"
	"time"

	"wklive/proto/staking"
	"wklive/services/system/internal/global"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

const stakingTaskRPCTimeout = 10 * time.Minute

func init() {
	cronx.Register("staking.ProcessRewardsAndSettleOrders", "质押收益发放/到期结算", runStakingProcessRewardsAndSettleOrders)
}

func runStakingProcessRewardsAndSettleOrders(ctx context.Context, job *models.SysJob) error {
	return callStakingTask(ctx, job, "process rewards and settle orders", global.StakingTaskCli.ProcessRewardsAndSettleOrders)
}

func callStakingTask(
	ctx context.Context,
	_ *models.SysJob,
	action string,
	fn func(context.Context, *staking.StakingTaskReq, ...grpcCallOption) (*staking.StakingTaskResp, error),
) error {
	rpcCtx, cancel := context.WithTimeout(ctx, stakingTaskRPCTimeout)
	defer cancel()

	result, err := fn(rpcCtx, &staking.StakingTaskReq{TenantId: 0})
	if err != nil {
		return err
	}
	if result == nil || result.Base == nil {
		return fmt.Errorf("staking %s failed, empty response", action)
	}
	if result.Base.Code != 200 {
		return fmt.Errorf("staking %s failed, code: %d, message: %s", action, result.Base.Code, result.Base.Msg)
	}
	return nil
}
