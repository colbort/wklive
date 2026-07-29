package tasks

import (
	"context"
	"fmt"

	"wklive/common/i18n"
	"wklive/common/tasks"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	logic "wklive/services/trade/internal/logic/task"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartTaskSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go helpers.RunSubscriberWithRestart(ctx, "trade task subscriber", func() error {
		return tasks.SubscribeService(ctx, svcCtx.TaskSubscriber, tasks.ServiceTrade, func(ctx context.Context, msg tasks.Message) error {
			logx.WithContext(ctx).Infof(
				"trade task received, action=%s jobId=%d tenantId=%d messageId=%s createdAt=%d",
				msg.Action, msg.JobID, msg.TenantID, msg.ID, msg.CreatedAt,
			)
			if err := handleTask(ctx, svcCtx, msg); err != nil {
				logx.WithContext(ctx).Errorf(
					"trade task failed, action=%s jobId=%d tenantId=%d messageId=%s err=%v",
					msg.Action, msg.JobID, msg.TenantID, msg.ID, err,
				)
				// Propagate the failure so the Kafka subscriber applies its
				// configured retry policy and eventually sends the message to
				// the DLQ. Returning nil here would acknowledge failed tasks.
				return err
			}
			logx.WithContext(ctx).Infof(
				"trade task completed, action=%s jobId=%d tenantId=%d messageId=%s",
				msg.Action, msg.JobID, msg.TenantID, msg.ID,
			)
			return nil
		})
	})
}

func handleTask(ctx context.Context, svcCtx *svc.ServiceContext, msg tasks.Message) error {
	req := &trade.TradeTaskReq{TenantId: msg.TenantID}
	handler := taskHandlerFor(msg.Action)
	if handler == nil {
		logx.Errorf("unknown trade task action: %s", msg.Action)
		return fmt.Errorf("unknown trade task action: %s", msg.Action)
	}
	return checkResp(handler(ctx, svcCtx, req))
}

type taskHandler func(context.Context, *svc.ServiceContext, *trade.TradeTaskReq) (*trade.TradeTaskResp, error)

func taskHandlerFor(action string) taskHandler {
	switch action {
	case tasks.ActionTradeProcessOrderMatching:
		return func(ctx context.Context, svcCtx *svc.ServiceContext, req *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
			return logic.NewProcessOrderMatchingLogic(ctx, svcCtx).ProcessOrderMatching(req)
		}
	case tasks.ActionTradeProcessPositions:
		return func(ctx context.Context, svcCtx *svc.ServiceContext, req *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
			return logic.NewProcessPositionsLogic(ctx, svcCtx).ProcessPositions(req)
		}
	case tasks.ActionTradeProcessContractSettlements:
		return func(ctx context.Context, svcCtx *svc.ServiceContext, req *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
			return logic.NewProcessContractSettlementsLogic(ctx, svcCtx).ProcessContractSettlements(req)
		}
	case tasks.ActionTradeProcessSecondsSettlements:
		return func(ctx context.Context, svcCtx *svc.ServiceContext, req *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
			return logic.NewProcessSecondsSettlementsLogic(ctx, svcCtx).ProcessSecondsSettlements(req)
		}
	case tasks.ActionTradeProcessTradeEvents:
		return func(ctx context.Context, svcCtx *svc.ServiceContext, req *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
			return logic.NewProcessTradeEventsLogic(ctx, svcCtx).ProcessTradeEvents(req)
		}
	case tasks.ActionTradeExpireRiskLimits:
		return func(ctx context.Context, svcCtx *svc.ServiceContext, req *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
			return logic.NewExpireRiskLimitsLogic(ctx, svcCtx).ExpireRiskLimits(req)
		}
	case tasks.ActionTradeArchiveLiquidityOrders:
		return func(ctx context.Context, svcCtx *svc.ServiceContext, req *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
			return logic.NewArchiveLiquidityOrdersLogic(ctx, svcCtx).ArchiveLiquidityOrders(req)
		}
	default:
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
	// Another scheduler or the in-process recovery safety net already owns the
	// distributed task lock. The work is being performed, so acknowledge this
	// Kafka message instead of retrying it and emitting misleading errors.
	if resp.GetBase().GetCode() == int32(i18n.SyncTaskAlreadyRunning) {
		return nil
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("task rejected, code=%d msg=%s", resp.GetBase().GetCode(), resp.GetBase().GetMsg())
	}
	return nil
}
