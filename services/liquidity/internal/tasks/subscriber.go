package tasks

import (
	"context"
	"fmt"

	"wklive/common/tasks"
	"wklive/proto/liquidity"
	tasklogic "wklive/services/liquidity/internal/logic/task"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartTaskSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		err := tasks.SubscribeService(ctx, svcCtx.TaskSubscriber, tasks.ServiceLiquidity, func(ctx context.Context, msg tasks.Message) error {
			if err := handleTask(ctx, svcCtx, msg); err != nil {
				logx.WithContext(ctx).Errorf("liquidity task failed, action=%s jobId=%d messageId=%s err=%v", msg.Action, msg.JobID, msg.ID, err)
				return err
			}
			return nil
		})
		if err != nil && ctx.Err() == nil {
			logx.Errorf("liquidity task subscriber stopped: %v", err)
		}
	}()
}

func handleTask(ctx context.Context, svcCtx *svc.ServiceContext, msg tasks.Message) error {
	req := &liquidity.LiquidityTaskReq{}
	switch msg.Action {
	case tasks.ActionLiquidityRefreshQuotes:
		return checkResp(tasklogic.NewRefreshQuotesLogic(ctx, svcCtx).RefreshQuotes(req))
	case tasks.ActionLiquidityRecoverQuoteOrders:
		return checkResp(tasklogic.NewRecoverQuoteOrdersLogic(ctx, svcCtx).RecoverQuoteOrders(req))
	default:
		logx.Errorf("unknown liquidity task action: %s", msg.Action)
		return nil
	}
}

func checkResp(resp *liquidity.LiquidityTaskResp, err error) error {
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
