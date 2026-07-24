package tasklogic

import (
	"context"
	"errors"
	"time"

	"wklive/proto/staking"
	"wklive/services/staking/internal/delayqueue"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartDelayQueue(ctx context.Context, svcCtx *svc.ServiceContext) {
	if svcCtx == nil || svcCtx.DelayQueue == nil {
		return
	}
	go svcCtx.DelayQueue.Consume(func(message delayqueue.Message) {
		if err := handleDelayMessage(ctx, svcCtx, message); err != nil {
			logx.WithContext(ctx).Errorf("staking delay message failed, orderId=%d err=%v", message.OrderID, err)
		}
	})
}

func handleDelayMessage(ctx context.Context, svcCtx *svc.ServiceContext, message delayqueue.Message) error {
	if message.Action != delayqueue.ActionMatureOrder {
		return nil
	}
	order, err := svcCtx.StakeOrderModel.FindOne(ctx, message.OrderID)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	if order.TenantId != message.TenantID || order.EndTimes != message.DueAt || now < order.EndTimes ||
		order.Status != int64(staking.OrderStatus_ORDER_STATUS_STAKING) {
		return nil
	}
	return NewProcessRewardsAndSettleOrdersLogic(ctx, svcCtx).settleExpiredOrder(order, now)
}
