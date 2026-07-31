package applogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	controlEventOrderRejected   = "ORDER_REJECTED"
	controlEventKillActivated   = "KILL_SWITCH_ACTIVATED"
	controlEventKillReleased    = "KILL_SWITCH_RELEASED"
	controlEventSTPPrevented    = "STP_PREVENTED"
	controlEventCircuitBreaker  = "CIRCUIT_BREAKER"
	controlReasonContractClosed = "CONTRACT_NOT_TRADABLE"
	controlReasonNotConfigured  = "CONTROL_NOT_CONFIGURED"
	controlReasonKillSwitch     = "USER_KILL_SWITCH"
	controlReasonStaleMark      = "PRICE_REFERENCE_STALE"
	controlReasonPriceBand      = "ORDER_PRICE_BAND"
	controlReasonUserLongLimit  = "USER_LONG_LIMIT"
	controlReasonUserShortLimit = "USER_SHORT_LIMIT"
	controlReasonOILimit        = "OPEN_INTEREST_LIMIT"
	controlReasonSelfTrade      = "SELF_TRADE_PREVENTED"
)

type orderControlRejection struct {
	reason string
	detail string
}

func evaluateOrderTradingControls(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	conn sqlx.SqlConn,
	order *models.TOptionOrder,
	now int64,
) (*models.TOptionContract, *orderControlRejection, error) {
	controlModel := models.NewTOptionUserTradingControlModel(conn, svcCtx.Config.CacheRedis)
	contractModel := models.NewTOptionContractModel(conn, svcCtx.Config.CacheRedis)
	eventModel := models.NewTOptionTradingControlEventModel(conn, svcCtx.Config.CacheRedis)
	marketModel := models.NewTOptionMarketModel(conn, svcCtx.Config.CacheRedis)
	positionModel := models.NewTOptionPositionModel(conn, svcCtx.Config.CacheRedis)
	orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
	outboxModel := models.NewTOptionOutboxModel(conn, svcCtx.Config.CacheRedis)
	haltModel := models.NewTOptionTradingHaltModel(conn, svcCtx.Config.CacheRedis)
	calendarModel := models.NewTOptionTradingCalendarModel(conn, svcCtx.Config.CacheRedis)
	calendarSessionModel := models.NewTOptionTradingCalendarSessionModel(conn, svcCtx.Config.CacheRedis)
	calendarExceptionModel := models.NewTOptionTradingCalendarExceptionModel(conn, svcCtx.Config.CacheRedis)

	contract, err := contractModel.FindOneForUpdate(ctx, order.ContractId)
	if err != nil {
		return nil, nil, err
	}
	userControl, err := controlModel.EnsureForUpdate(ctx, order.TenantId, order.UserId, now)
	if err != nil {
		return nil, nil, err
	}
	reject := func(reason, detail string) (*models.TOptionContract, *orderControlRejection, error) {
		if _, insertErr := eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: order.TenantId, UserId: order.UserId, ContractId: order.ContractId,
			EventType: controlEventOrderRejected, Reason: reason, Detail: detail,
			OperatorId: order.UserId, CreateTimes: now,
		}); insertErr != nil {
			return nil, nil, insertErr
		}
		return contract, &orderControlRejection{reason: reason, detail: detail}, nil
	}

	if contract.TenantId != order.TenantId ||
		contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
		contract.IsDeleted == int64(common.YesNo_YES_NO_YES) ||
		now < contract.ListTime || (contract.ExpireTime > 0 && now >= contract.ExpireTime) {
		return reject(controlReasonContractClosed, "contract is not in its tradable window")
	}
	calendarDecision, calendarErr := logichelpers.IsContractTradingOpenWithModels(
		ctx, haltModel, calendarModel, calendarSessionModel, calendarExceptionModel, contract, now,
	)
	if calendarErr != nil {
		return reject(controlReasonContractClosed, "trading calendar evaluation failed")
	}
	if calendarDecision == nil || !calendarDecision.Open {
		reason := "trading calendar is closed"
		if calendarDecision != nil && calendarDecision.Reason != "" {
			reason = calendarDecision.Reason
		}
		return reject(controlReasonContractClosed, reason)
	}
	if userControl.KillSwitch == int64(common.YesNo_YES_NO_YES) {
		return reject(controlReasonKillSwitch, fmt.Sprintf("activated_at=%d", userControl.ActivatedAt))
	}
	if !contract.MaxUserLongQty.IsPositive() || !contract.MaxUserShortQty.IsPositive() ||
		!contract.MaxOpenInterest.IsPositive() || !contract.OrderPriceBandRatio.IsPositive() ||
		!contract.CircuitBreakerRatio.IsPositive() {
		return reject(controlReasonNotConfigured, "one or more mandatory trading controls are zero")
	}
	market, err := marketModel.FindOneByTenantIdContractId(ctx, order.TenantId, order.ContractId)
	if err != nil || !logichelpers.IsMarkFresh(market, now, 30) || !market.MarkPrice.IsPositive() {
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, nil, err
		}
		return reject(controlReasonStaleMark, "fresh positive mark price is required")
	}
	lower, upper, withinBand := optionOrderPriceBand(
		order.Price, market.MarkPrice, contract.OrderPriceBandRatio,
	)
	if !withinBand {
		return reject(controlReasonPriceBand, fmt.Sprintf(
			"price=%s mark=%s lower=%s upper=%s",
			order.Price, market.MarkPrice, lower, upper,
		))
	}
	if order.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_OPEN) {
		if mmpRejection, err := ensureMMPAdmission(ctx, svcCtx, conn, order, now); err != nil {
			return nil, nil, err
		} else if mmpRejection != nil {
			return reject(mmpRejection.reason, mmpRejection.detail)
		}
		return contract, nil, nil
	}

	positionSide := int64(common.PositionSide_POSITION_SIDE_LONG)
	userLimit := contract.MaxUserLongQty
	userLimitReason := controlReasonUserLongLimit
	if order.Side == int64(common.Side_SIDE_SELL) {
		positionSide = int64(common.PositionSide_POSITION_SIDE_SHORT)
		userLimit = contract.MaxUserShortQty
		userLimitReason = controlReasonUserShortLimit
	}
	userExposure, err := projectedOptionExposure(
		ctx, positionModel, orderModel, outboxModel,
		order.TenantId, order.UserId, order.ContractId, positionSide,
	)
	if err != nil {
		return nil, nil, err
	}
	if optionExposureLimitExceeded(userExposure, order.Qty, userLimit) {
		return reject(userLimitReason, fmt.Sprintf(
			"exposure=%s request=%s limit=%s side=%d",
			userExposure, order.Qty, userLimit, positionSide,
		))
	}
	contractExposure, err := projectedOptionExposure(
		ctx, positionModel, orderModel, outboxModel,
		order.TenantId, 0, order.ContractId, positionSide,
	)
	if err != nil {
		return nil, nil, err
	}
	if optionExposureLimitExceeded(contractExposure, order.Qty, contract.MaxOpenInterest) {
		return reject(controlReasonOILimit, fmt.Sprintf(
			"exposure=%s request=%s limit=%s side=%d",
			contractExposure, order.Qty, contract.MaxOpenInterest, positionSide,
		))
	}
	if mmpRejection, err := ensureMMPAdmission(ctx, svcCtx, conn, order, now); err != nil {
		return nil, nil, err
	} else if mmpRejection != nil {
		return reject(mmpRejection.reason, mmpRejection.detail)
	}
	return contract, nil, nil
}

func optionOrderPriceBand(
	price, markPrice, ratio decimal.Decimal,
) (decimal.Decimal, decimal.Decimal, bool) {
	if !price.IsPositive() || !markPrice.IsPositive() || !ratio.IsPositive() ||
		ratio.GreaterThan(decimal.NewFromInt(1)) {
		return decimal.Zero, decimal.Zero, false
	}
	lower := decimal.Max(
		markPrice.Mul(decimal.NewFromInt(1).Sub(ratio)),
		decimal.Zero,
	)
	upper := markPrice.Mul(decimal.NewFromInt(1).Add(ratio))
	return lower, upper, !price.LessThan(lower) && !price.GreaterThan(upper)
}

func optionExposureLimitExceeded(exposure, request, limit decimal.Decimal) bool {
	return !limit.IsPositive() || request.IsNegative() || exposure.Add(request).GreaterThan(limit)
}

func projectedOptionExposure(
	ctx context.Context,
	positionModel models.TOptionPositionModel,
	orderModel models.TOptionOrderModel,
	outboxModel models.TOptionOutboxModel,
	tenantID, userID, contractID, positionSide int64,
) (decimal.Decimal, error) {
	positionQty, err := positionModel.SumHoldingQty(ctx, tenantID, userID, contractID, positionSide)
	if err != nil {
		return decimal.Zero, err
	}
	orderSide := int64(common.Side_SIDE_BUY)
	if positionSide == int64(common.PositionSide_POSITION_SIDE_SHORT) {
		orderSide = int64(common.Side_SIDE_SELL)
	}
	activeOpenQty, err := orderModel.SumActiveOpenQty(ctx, tenantID, userID, contractID, orderSide)
	if err != nil {
		return decimal.Zero, err
	}
	pendingDelta, err := outboxModel.SumPendingPositionDelta(ctx, tenantID, userID, contractID, positionSide)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.Max(positionQty.Add(activeOpenQty).Add(pendingDelta), decimal.Zero), nil
}

func insertTradingControlEvent(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	conn sqlx.SqlConn,
	event *models.TOptionTradingControlEvent,
) error {
	if event.CreateTimes == 0 {
		event.CreateTimes = time.Now().Unix()
	}
	if _, err := models.NewTOptionTradingControlEventModel(
		conn, svcCtx.Config.CacheRedis,
	).Insert(ctx, event); err != nil {
		return err
	}
	logx.WithContext(ctx).Infof(
		"option trading control metric event=%s reason=%s tenantId=%d userId=%d contractId=%d orderId=%d",
		event.EventType, event.Reason, event.TenantId, event.UserId, event.ContractId, event.OrderId,
	)
	return nil
}

func optionUserTradingControl(item *models.TOptionUserTradingControl) *option.OptionUserTradingControl {
	if item == nil {
		return &option.OptionUserTradingControl{KillSwitch: common.YesNo_YES_NO_NO}
	}
	return &option.OptionUserTradingControl{
		TenantId: item.TenantId, UserId: item.UserId,
		KillSwitch: common.YesNo(item.KillSwitch), Reason: item.Reason,
		ActivatedAt: item.ActivatedAt, ReleasedAt: item.ReleasedAt, UpdateTimes: item.UpdateTimes,
	}
}
