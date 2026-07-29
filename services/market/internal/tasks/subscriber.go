package tasks

import (
	"context"
	"fmt"

	"wklive/common/tasks"
	"wklive/proto/market"
	logic "wklive/services/market/internal/logic/task"
	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartTaskSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		if err := tasks.SubscribeService(ctx, svcCtx.TaskSubscriber, tasks.ServiceMarket, func(ctx context.Context, msg tasks.Message) error {
			if err := handleTask(ctx, svcCtx, msg); err != nil {
				logx.Errorf("market task failed, action=%s jobId=%d err=%v", msg.Action, msg.JobID, err)
			}
			return nil
		}); err != nil && ctx.Err() == nil {
			logx.Errorf("market task subscriber stopped: %v", err)
		}
	}()
}

func handleTask(ctx context.Context, svcCtx *svc.ServiceContext, msg tasks.Message) error {
	switch msg.Action {
	case tasks.ActionMarketSyncProducts:
		return checkSyncProductsResp(logic.NewSyncProductsLogic(ctx, svcCtx).SyncProducts(&market.SyncProductsReq{}))
	case tasks.ActionMarketSyncKlines:
		return checkSyncKlinesResp(logic.NewSyncKlinesLogic(ctx, svcCtx).SyncKlines(&market.SyncKlinesReq{}))
	default:
		logx.Errorf("unknown market task action: %s", msg.Action)
		return nil
	}
}

func checkSyncProductsResp(resp *market.SyncProductsResp, err error) error {
	if err != nil {
		return err
	}
	if resp == nil || resp.GetBase() == nil {
		return fmt.Errorf("empty sync products response")
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("sync products rejected, code=%d msg=%s", resp.GetBase().GetCode(), resp.GetBase().GetMsg())
	}
	return nil
}

func checkSyncKlinesResp(resp *market.SyncKlinesResp, err error) error {
	if err != nil {
		return err
	}
	if resp == nil || resp.GetBase() == nil {
		return fmt.Errorf("empty sync klines response")
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("sync klines rejected, code=%d msg=%s", resp.GetBase().GetCode(), resp.GetBase().GetMsg())
	}
	return nil
}
