package adminlogic

import (
	"context"
	"errors"
	"wklive/proto/common"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func beginSystemOrderTermination(ctx context.Context, svcCtx *svc.ServiceContext, orderID int64, reason string, rejectReduceOnly bool) (*models.TTradeOrder, error) {
	var terminating *models.TTradeOrder
	err := helpers.TransactWithDeadlockRetry(ctx, svcCtx.DB, func(txCtx context.Context, session sqlx.Session) error {
		orderModel := models.NewTTradeOrderModel(sqlx.NewSqlConnFromSession(session), svcCtx.Config.CacheRedis)
		current, err := orderModel.FindOneForUpdate(txCtx, orderID)
		if err != nil {
			return err
		}
		if !helpers.IsOpenOrderStatus(current.Status) || rejectReduceOnly && current.IsReduceOnly == 1 {
			return nil
		}
		current.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELING)
		current.CanceledQty = decimalMaxZero(current.Qty.Sub(current.FilledQty))
		current.CancelReason = reason
		current.Version++
		current.UpdateTimes = utils.NowMillis()
		if err = orderModel.Update(txCtx, current); err != nil {
			return err
		}
		terminating = current
		return nil
	})
	return terminating, err
}

func finalizeOrderTermination(ctx context.Context, conn sqlx.SqlConn, svcCtx *svc.ServiceContext, orderID int64, now int64) (bool, error) {
	orderModel := models.NewTTradeOrderModel(conn, svcCtx.Config.CacheRedis)
	reservationModel := models.NewTTradeAssetReservationModel(conn, svcCtx.Config.CacheRedis)
	instructionModel := models.NewTTradeSettlementInstructionModel(conn, svcCtx.Config.CacheRedis)
	fillModel := models.NewTTradeFillModel(conn, svcCtx.Config.CacheRedis)
	contractOrderModel := models.NewTTradeOrderContractModel(conn, svcCtx.Config.CacheRedis)
	positionModel := models.NewTContractPositionModel(conn, svcCtx.Config.CacheRedis)
	eventModel := models.NewTBizTradeEventModel(conn, svcCtx.Config.CacheRedis)

	order, err := orderModel.FindOneForUpdate(ctx, orderID)
	if err != nil {
		return false, err
	}
	status := trade.OrderStatus(order.Status)
	if status != trade.OrderStatus_ORDER_STATUS_CANCELING && status != trade.OrderStatus_ORDER_STATUS_EXPIRING {
		return false, nil
	}
	reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, order.TenantId, order.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return false, err
	}
	if reservation != nil && reservation.ReservedAmount.Sub(reservation.ConsumedAmount).Sub(reservation.ReleasedAmount).IsPositive() {
		return false, nil
	}
	unfinished, err := instructionModel.CountAllUnfinishedByOrder(ctx, order.TenantId, order.Id)
	if err != nil || unfinished > 0 {
		return false, err
	}
	unsettledFills, err := fillModel.CountUnsettledByOrder(ctx, order.TenantId, order.Id)
	if err != nil || unsettledFills > 0 {
		return false, err
	}
	if order.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		contractOrder, findErr := contractOrderModel.FindOneByTenantIdOrderId(ctx, order.TenantId, order.Id)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return false, findErr
		}
		if contractOrder != nil && contractOrder.ReservedCloseQty.IsPositive() {
			lookupSide := order.PositionSide
			if lookupSide == int64(trade.PositionSide_POSITION_SIDE_NET) {
				lookupSide, _ = netPositionSides(order.Side)
			}
			position, findErr := positionModel.FindOneByTenantIdUserIdSymbolIdPositionSideMarginMode(ctx, order.TenantId, order.UserId, order.SymbolId, lookupSide, contractOrder.MarginMode)
			if findErr != nil {
				return false, findErr
			}
			if err := positionModel.ReleaseCloseQty(ctx, position.Id, contractOrder.ReservedCloseQty, now); err != nil {
				return false, err
			}
			contractOrder.ReservedCloseQty = contractOrder.ReservedCloseQty.Sub(contractOrder.ReservedCloseQty)
			contractOrder.UpdateTimes = now
			if err := contractOrderModel.Update(ctx, contractOrder); err != nil {
				return false, err
			}
		}
	}
	finalStatus := trade.OrderStatus_ORDER_STATUS_CANCELED
	eventType, suffix := "ORDER_CANCELED", "CANCELED"
	if status == trade.OrderStatus_ORDER_STATUS_EXPIRING {
		finalStatus, eventType, suffix = trade.OrderStatus_ORDER_STATUS_EXPIRED, "ORDER_EXPIRED", "EXPIRED"
	}
	order.Status = int64(finalStatus)
	order.Version++
	order.UpdateTimes = now
	if err := orderModel.Update(ctx, order); err != nil {
		return false, err
	}
	if err := insertMatchOutboxEvent(ctx, eventModel, order, derivedTradeBizNo(order.OrderNo, suffix), eventType, order.OrderNo, "order", "{}", now); err != nil {
		return false, err
	}
	return true, nil
}

func finalizeSettledOrder(ctx context.Context, conn sqlx.SqlConn, svcCtx *svc.ServiceContext, orderID int64, now int64) error {
	orderModel := models.NewTTradeOrderModel(conn, svcCtx.Config.CacheRedis)
	reservationModel := models.NewTTradeAssetReservationModel(conn, svcCtx.Config.CacheRedis)
	instructionModel := models.NewTTradeSettlementInstructionModel(conn, svcCtx.Config.CacheRedis)
	fillModel := models.NewTTradeFillModel(conn, svcCtx.Config.CacheRedis)
	eventModel := models.NewTBizTradeEventModel(conn, svcCtx.Config.CacheRedis)
	order, err := orderModel.FindOneForUpdate(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING) {
		return nil
	}
	unfinished, err := instructionModel.CountAllUnfinishedByOrder(ctx, order.TenantId, order.Id)
	if err != nil || unfinished > 0 {
		return err
	}
	unsettledFills, err := fillModel.CountUnsettledByOrder(ctx, order.TenantId, order.Id)
	if err != nil || unsettledFills > 0 {
		return err
	}
	reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, order.TenantId, order.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if reservation != nil && reservation.ReservedAmount.Sub(reservation.ConsumedAmount).Sub(reservation.ReleasedAmount).IsPositive() {
		return nil
	}
	order.Status = int64(trade.OrderStatus_ORDER_STATUS_FILLED)
	order.CompletionReason = "FILLED_AND_SETTLED"
	order.Version++
	order.UpdateTimes = now
	if err := orderModel.Update(ctx, order); err != nil {
		return err
	}
	return insertMatchOutboxEvent(ctx, eventModel, order, derivedTradeBizNo(order.OrderNo, "SETTLED"), "ORDER_SETTLED", order.OrderNo, "order", "{}", now)
}
