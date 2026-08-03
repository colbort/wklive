package tasks

import (
	"context"
	"fmt"

	"wklive/common/tasks"
	"wklive/proto/staking"
	logic "wklive/services/staking/internal/logic/task"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartTaskSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		if err := tasks.SubscribeService(ctx, svcCtx.TaskSubscriber, tasks.ServiceStaking, func(ctx context.Context, msg tasks.Message) error {
			if err := handleTask(ctx, svcCtx, msg); err != nil {
				logx.Errorf("staking task failed, action=%s jobId=%d err=%v", msg.Action, msg.JobID, err)
			}
			return nil
		}); err != nil && ctx.Err() == nil {
			logx.Errorf("staking task subscriber stopped: %v", err)
		}
	}()
}

func handleTask(ctx context.Context, svcCtx *svc.ServiceContext, msg tasks.Message) error {
	switch msg.Action {
	case tasks.ActionStakingProcessRewardsAndSettleOrders:
		req := &staking.StakingTaskReq{TenantId: msg.TenantID}
		return checkResp(logic.NewProcessRewardsAndSettleOrdersLogic(ctx, svcCtx).ProcessRewardsAndSettleOrders(req))
	case tasks.ActionStakingReconcile:
		req := &staking.StakingTaskReq{TenantId: msg.TenantID}
		return checkResp(logic.NewReconcileStakingLogic(ctx, svcCtx).ReconcileStaking(req))
	default:
		logx.Errorf("unknown staking task action: %s", msg.Action)
		return nil
	}
}

func checkResp(resp *staking.StakingTaskResp, err error) error {
	if err != nil {
		return err
	}
	if resp == nil || resp.GetBase() == nil {
		return fmt.Errorf("empty task response")
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("task rejected, code=%d msg=%s", resp.GetBase().GetCode(), resp.GetBase().GetMsg())
	}
	return nil
}
