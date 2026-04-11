package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAllOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelAllOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAllOrdersLogic {
	return &CancelAllOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 撤销当前用户全部订单
func (l *CancelAllOrdersLogic) CancelAllOrders(in *trade.CancelAllOrdersReq) (*trade.CancelAllOrdersResp, error) {
	list, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
		TenantId:     int64(in.TenantId),
		UserId:       int64(in.UserId),
		SymbolId:     int64(in.SymbolId),
		MarketType:   int64(in.MarketType),
		Side:         int64(in.Side),
		Statuses:     openOrderStatuses(),
		PositionSide: int64(in.PositionSide),
	}, 0, 100)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	affected := uint32(0)
	for _, item := range list {
		item.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELED)
		item.CancelReason = orderCancelReason("user")
		item.UpdateTimes = utils.NowMillis()
		if err = l.svcCtx.TradeOrderModel.Update(l.ctx, item); err != nil {
			return nil, err
		}
		if _, err = l.svcCtx.TradeCancelLogModel.Insert(l.ctx, &models.TTradeCancelLog{
			TenantId:     item.TenantId,
			OrderId:      item.Id,
			OrderNo:      item.OrderNo,
			UserId:       item.UserId,
			CancelSource: int64(trade.CancelSource_CANCEL_SOURCE_USER),
			CancelReason: item.CancelReason,
			CreateTimes:  utils.NowMillis(),
		}); err != nil {
			return nil, err
		}
		affected++
	}

	return &trade.CancelAllOrdersResp{Base: helper.OkResp(), AffectedCount: affected}, nil
}
