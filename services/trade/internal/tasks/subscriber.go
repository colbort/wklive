package tasks

import (
	"context"
	"fmt"

	"wklive/common/tasks"
	"wklive/proto/trade"
	logic "wklive/services/trade/internal/logic/tradetask"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartTaskSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		if err := tasks.SubscribeService(ctx, svcCtx.TaskSubscriber, tasks.ServiceTrade, func(ctx context.Context, msg tasks.Message) error {
			if err := handleTask(ctx, svcCtx, msg); err != nil {
				logx.Errorf("trade task failed, action=%s jobId=%d err=%v", msg.Action, msg.JobID, err)
			}
			return nil
		}); err != nil && ctx.Err() == nil {
			logx.Errorf("trade task subscriber stopped: %v", err)
		}
	}()
}

func handleTask(ctx context.Context, svcCtx *svc.ServiceContext, msg tasks.Message) error {
	req := &trade.TradeTaskReq{TenantId: msg.TenantID}

	switch msg.Action {
	case tasks.ActionTradeProcessOrderMatching:
		return checkResp(logic.NewProcessOrderMatchingLogic(ctx, svcCtx).ProcessOrderMatching(req))
	case tasks.ActionTradeProcessPositions:
		return checkResp(logic.NewProcessPositionsLogic(ctx, svcCtx).ProcessPositions(req))
	case tasks.ActionTradeProcessContractSettlements:
		return checkResp(logic.NewProcessContractSettlementsLogic(ctx, svcCtx).ProcessContractSettlements(req))
	case tasks.ActionTradeProcessTradeEvents:
		return checkResp(logic.NewProcessTradeEventsLogic(ctx, svcCtx).ProcessTradeEvents(req))
	case tasks.ActionTradeExpireRiskLimits:
		return checkResp(logic.NewExpireRiskLimitsLogic(ctx, svcCtx).ExpireRiskLimits(req))
	default:
		logx.Errorf("unknown trade task action: %s", msg.Action)
		return nil
	}
}

func checkResp(resp *trade.TradeTaskResp, err error) error {
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
