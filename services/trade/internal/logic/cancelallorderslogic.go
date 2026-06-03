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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	affected := int64(0)
	cursor := int64(0)
	for {
		list, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
			TenantId:     tenantId,
			UserId:       userId,
			SymbolId:     in.SymbolId,
			MarketType:   int64(in.MarketType),
			Side:         int64(in.Side),
			Statuses:     openOrderStatuses(),
			PositionSide: int64(in.PositionSide),
		}, cursor, 100)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if len(list) == 0 {
			break
		}
		for _, item := range list {
			cursor = item.Id
			var canceledOrder *models.TTradeOrder
			if err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				conn := sqlx.NewSqlConnFromSession(session)
				orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeOrderModel)
				cancelLogModel := models.NewTTradeCancelLogModel(conn, l.svcCtx.Config.CacheRedis).(models.TradeCancelLogModel)

				locked, err := orderModel.FindOneForUpdate(ctx, item.Id)
				if err != nil {
					return err
				}
				if locked.TenantId != tenantId || locked.UserId != userId || !isOpenOrderStatus(locked.Status) {
					return nil
				}
				locked.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELED)
				locked.CancelReason = orderCancelReason("user")
				locked.UpdateTimes = utils.NowMillis()
				if err := orderModel.Update(ctx, locked); err != nil {
					return err
				}
				_, err = cancelLogModel.Insert(ctx, &models.TTradeCancelLog{
					TenantId:     locked.TenantId,
					OrderId:      locked.Id,
					OrderNo:      locked.OrderNo,
					UserId:       locked.UserId,
					CancelSource: int64(trade.CancelSource_CANCEL_SOURCE_USER),
					CancelReason: locked.CancelReason,
					CreateTimes:  utils.NowMillis(),
				})
				if err != nil {
					return err
				}
				canceledOrder = locked
				return nil
			}); err != nil {
				return nil, err
			}
			if canceledOrder != nil {
				if err = removeOrderBookOrder(l.svcCtx, l.ctx, canceledOrder); err != nil {
					return nil, err
				}
				if err = unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, canceledOrder, "trade cancel all orders unfreeze"); err != nil {
					return nil, err
				}
				affected++
			}
		}
		if len(list) < 100 {
			break
		}
	}

	return &trade.CancelAllOrdersResp{Base: helper.OkResp(), AffectedCount: affected}, nil
}
