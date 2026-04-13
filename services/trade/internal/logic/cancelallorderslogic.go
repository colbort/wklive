package logic

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
		ext, parseErr := parseOrderAssetExt(conv.NullStringValue(item.BizExt))
		if parseErr != nil {
			return nil, parseErr
		}
		if ext.FreezeNo != "" {
			var unfreezeAmount float64
			if item.MarketType == int64(trade.MarketType_MARKET_TYPE_SPOT) {
				spot, findErr := l.svcCtx.TradeOrderSpotModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
				if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
					return nil, findErr
				}
				if spot != nil {
					unfreezeAmount = spot.FrozenAmount
				}
			} else {
				contract, findErr := l.svcCtx.TradeOrderContractModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
				if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
					return nil, findErr
				}
				if contract != nil {
					unfreezeAmount = contract.MarginAmount
				}
			}
			if err = unfreezeOrderAsset(l.svcCtx, l.ctx, item, ext.FreezeNo, unfreezeAmount, "trade cancel all orders unfreeze"); err != nil {
				return nil, err
			}
		}

		item.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELED)
		item.CancelReason = orderCancelReason("user")
		item.UpdateTimes = utils.NowMillis()
		if err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			conn := sqlx.NewSqlConnFromSession(session)
			orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeOrderModel)
			cancelLogModel := models.NewTTradeCancelLogModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeCancelLogModel)

			if err := orderModel.Update(ctx, item); err != nil {
				return err
			}
			_, err := cancelLogModel.Insert(ctx, &models.TTradeCancelLog{
				TenantId:     item.TenantId,
				OrderId:      item.Id,
				OrderNo:      item.OrderNo,
				UserId:       item.UserId,
				CancelSource: int64(trade.CancelSource_CANCEL_SOURCE_USER),
				CancelReason: item.CancelReason,
				CreateTimes:  utils.NowMillis(),
			})
			return err
		}); err != nil {
			return nil, err
		}
		affected++
	}

	return &trade.CancelAllOrdersResp{Base: helper.OkResp(), AffectedCount: affected}, nil
}
