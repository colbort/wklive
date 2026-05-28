package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessTradeEventsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessTradeEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessTradeEventsLogic {
	return &ProcessTradeEventsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 交易事件处理（失败重试/冻结资产修复）
func (l *ProcessTradeEventsLogic) ProcessTradeEvents(in *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
	return runTradeTaskWithLock(l.ctx, l.svcCtx, "process_trade_events", func() (*trade.TradeTaskResp, error) {
		if err := l.retryTradeEvents(in); err != nil {
			return nil, err
		}
		if err := l.repairFrozenAssets(in); err != nil {
			return nil, err
		}
		return okTradeTaskResp(), nil
	})
}

func (l *ProcessTradeEventsLogic) retryTradeEvents(in *trade.TradeTaskReq) error {
	now := utils.NowMillis()
	cursor := int64(0)
	for {
		items, _, err := l.svcCtx.BizTradeEventModel.FindPage(l.ctx, models.BizTradeEventPageFilter{
			TenantId:    in.GetTenantId(),
			EventStatus: int64(trade.EventStatus_EVENT_STATUS_FAILED),
			TimeEnd:     now,
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for _, item := range items {
			cursor = item.Id
			if item.MaxRetryCount > 0 && item.RetryCount >= item.MaxRetryCount {
				continue
			}
			if item.NextRetryAt > 0 && item.NextRetryAt > now {
				continue
			}
			item.EventStatus = int64(trade.EventStatus_EVENT_STATUS_PENDING)
			item.RetryCount++
			item.NextRetryAt = now
			item.LastErrorMsg = ""
			item.UpdateTimes = now
			if err := l.svcCtx.BizTradeEventModel.Update(l.ctx, item); err != nil {
				return err
			}
		}
		if len(items) < 100 {
			return nil
		}
	}
}

func (l *ProcessTradeEventsLogic) repairFrozenAssets(in *trade.TradeTaskReq) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
			TenantId: in.GetTenantId(),
			Statuses: []int64{
				int64(trade.OrderStatus_ORDER_STATUS_CANCELED),
				int64(trade.OrderStatus_ORDER_STATUS_REJECTED),
				int64(trade.OrderStatus_ORDER_STATUS_EXPIRED),
			},
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, order := range orders {
			cursor = order.Id
			if err := createTradeTaskEvent(l.ctx, l.svcCtx, order.TenantId, "FROZEN_ASSET_REPAIR_REQUIRED", "order", order.Id, order.UserId, order.SymbolId, order.MarketType, "frozen asset repair task"); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}
