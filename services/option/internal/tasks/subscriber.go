package tasks

import (
	"context"
	"fmt"

	"wklive/common/tasks"
	"wklive/proto/option"
	logic "wklive/services/option/internal/logic/optiontask"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartTaskSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		if err := tasks.SubscribeService(ctx, svcCtx.TaskSubscriber, tasks.ServiceOption, func(ctx context.Context, msg tasks.Message) error {
			if err := handleTask(ctx, svcCtx, msg); err != nil {
				logx.Errorf("option task failed, action=%s jobId=%d err=%v", msg.Action, msg.JobID, err)
			}
			return nil
		}); err != nil && ctx.Err() == nil {
			logx.Errorf("option task subscriber stopped: %v", err)
		}
	}()
}

func handleTask(ctx context.Context, svcCtx *svc.ServiceContext, msg tasks.Message) error {
	req := &option.OptionTaskReq{TenantId: msg.TenantID}

	switch msg.Action {
	case tasks.ActionOptionProcessContractLifecycle:
		return checkResp(logic.NewProcessContractLifecycleLogic(ctx, svcCtx).ProcessContractLifecycle(req))
	case tasks.ActionOptionCleanMarketSnapshots:
		return checkResp(logic.NewCleanMarketSnapshotsLogic(ctx, svcCtx).CleanMarketSnapshots(req))
	default:
		logx.Errorf("unknown option task action: %s", msg.Action)
		return nil
	}
}

func checkResp(resp *option.OptionTaskResp, err error) error {
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
