package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	spotSettlementBatchSize = int64(100)
	spotSettlementMaxSteps  = 1000
	spotSettlementMaxRetry  = int64(20)
)

type ProcessFillSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessFillSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessFillSettlementsLogic {
	return &ProcessFillSettlementsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ProcessFill is the real-time entry point invoked by FILL_CREATED. It handles
// only one Fill and preserves step ordering; Process remains the recovery scan.
func (l *ProcessFillSettlementsLogic) ProcessFill(fillID int64) error {
	fill, err := l.svcCtx.TradeFillModel.FindOne(l.ctx, fillID)
	if err != nil {
		return err
	}
	if fill.SettlementStatus == int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_SETTLED) {
		return nil
	}
	for step := 0; step < 16; step++ {
		items, err := l.svcCtx.TradeSettlementInstrModel.FindByFillId(l.ctx, fill.TenantId, fill.Id)
		if err != nil {
			return err
		}
		var next *models.TTradeSettlementInstruction
		for _, item := range items {
			if item.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
				continue
			}
			if item.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW) {
				return nil
			}
			next = item
			break
		}
		if next == nil {
			return l.settleFillIfReady(fill)
		}
		if isUnboundProjectedMarginInstruction(next) {
			if err := NewProcessContractPositionFillsLogic(l.ctx, l.svcCtx).ProcessFill(next.FillId); err != nil {
				return err
			}
			continue
		}
		now := utils.NowMillis()
		if next.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED) && next.NextRetryAt > now || next.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PROCESSING) && next.UpdateTimes > now-60*1000 {
			return nil
		}
		claimed, lease, err := l.svcCtx.TradeSettlementInstrModel.ClaimLease(l.ctx, next.Id, now)
		if err != nil {
			return err
		}
		if !claimed {
			return nil
		}
		next.UpdateTimes = lease
		if err := l.processInstruction(next); err != nil {
			if markErr := l.markFailed(next, err); markErr != nil {
				return markErr
			}
			return err
		}
	}
	return fmt.Errorf("spot fill settlement exceeded maximum steps: %d", fillID)
}

// Process scans pending instructions as a recovery path for lost events,
// expired processing leases and retryable failures.
func (l *ProcessFillSettlementsLogic) Process(tenantID int64) error {
	for processed := 0; processed < spotSettlementMaxSteps; {
		now := utils.NowMillis()
		items, err := l.svcCtx.TradeSettlementInstrModel.FindPendingFillSettlements(l.ctx, tenantID, now, spotSettlementBatchSize)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		progressed := false
		for _, item := range items {
			if isUnboundProjectedMarginInstruction(item) {
				if err := NewProcessContractPositionFillsLogic(l.ctx, l.svcCtx).ProcessFill(item.FillId); err != nil {
					return err
				}
				progressed = true
				processed++
				continue
			}
			claimed, lease, err := l.svcCtx.TradeSettlementInstrModel.ClaimLease(l.ctx, item.Id, now)
			if err != nil {
				return err
			}
			if !claimed {
				continue
			}
			item.UpdateTimes = lease
			progressed = true
			processed++
			if err := l.processInstruction(item); err != nil {
				l.Errorf("spot settlement instruction failed, instructionNo=%s fillId=%d orderId=%d err=%v", item.InstructionNo, item.FillId, item.OrderId, err)
				if markErr := l.markFailed(item, err); markErr != nil {
					return markErr
				}
			}
		}
		if !progressed || int64(len(items)) < spotSettlementBatchSize {
			// Query again because completing step N may make step N+1 eligible.
			continue
		}
	}
	return nil
}

func isUnboundProjectedMarginInstruction(item *models.TTradeSettlementInstruction) bool {
	return item != nil &&
		item.Action == int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_ADJUST_MARGIN) &&
		item.PositionId == 0 &&
		item.FillId > 0
}

func (l *ProcessFillSettlementsLogic) processInstruction(item *models.TTradeSettlementInstruction) error {
	fill, err := l.svcCtx.TradeFillModel.FindOne(l.ctx, item.FillId)
	if err != nil {
		return err
	}
	order, err := l.svcCtx.TradeOrderModel.FindOne(l.ctx, item.OrderId)
	if err != nil {
		return err
	}
	if fill.TenantId != item.TenantId || order.TenantId != item.TenantId || fill.OrderId != order.Id || fill.ProductType != order.ProductType {
		return fmt.Errorf("settlement instruction does not match fill")
	}
	if fill.SettlementStatus != int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_PROCESSING) {
		fill.SettlementStatus = int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_PROCESSING)
		if err := l.svcCtx.TradeFillModel.Update(l.ctx, fill); err != nil {
			return err
		}
	}
	if err := l.executeAssetInstruction(item, fill, order); err != nil {
		return err
	}
	return l.markSucceeded(item, fill, order)
}

func (l *ProcessFillSettlementsLogic) executeAssetInstruction(item *models.TTradeSettlementInstruction, fill *models.TTradeFill, order *models.TTradeOrder) error {
	walletType := helpers.WalletTypeForProduct(common.ProductType(fill.ProductType))
	matchReq := func(scene asset.SceneType) (asset.BizType, asset.SceneType, int64, string) {
		return asset.BizType_BIZ_TYPE_TRADE, scene, fill.Id, item.InstructionNo
	}
	var resp *asset.ChangeAssetResp
	var err error
	switch trade.SettlementInstructionAction(item.Action) {
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.DeductFrozenAssetByBizNo(l.ctx, &asset.DeductFrozenAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill consume frozen"})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.UnfreezeAssetByBizNo(l.ctx, &asset.UnfreezeAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill release remainder"})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType, Coin: item.Asset, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill credit"})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_FEE)
		if fill.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) || order.Side == int64(common.Side_SIDE_BUY) {
			resp, err = l.svcCtx.AssetClient.DeductFrozenAssetByBizNo(l.ctx, &asset.DeductFrozenAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill fee from frozen"})
			if i18n.IsStatusError(err, i18n.FreezeRecordNotDeductible) {
				// Legacy residual cancellation could release the reservation
				// before its fee instruction ran. The same biz number keeps
				// this fallback idempotent in Asset, so a previously applied
				// fee can never be charged twice.
				resp, err = l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{
					TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType,
					Coin: item.Asset, Amount: item.Amount.String(),
					BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo,
					Remark: "spot fill fee recovery from available",
				})
			}
		} else {
			resp, err = l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType, Coin: item.Asset, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill fee from proceeds"})
		}
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_ADJUST_MARGIN:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.DeductFrozenAssetByBizNo(l.ctx, &asset.DeductFrozenAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "contract fill consume margin"})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_MARGIN, trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_POST_PNL:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType, Coin: item.Asset, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "contract margin or profit credit"})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType, Coin: item.Asset, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "contract realized loss debit"})
	default:
		return fmt.Errorf("unsupported spot settlement action: %d", item.Action)
	}
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(l.ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(l.ctx, resp.Base.Code)
	}
	return nil
}

func (l *ProcessFillSettlementsLogic) markSucceeded(item *models.TTradeSettlementInstruction, fill *models.TTradeFill, order *models.TTradeOrder) error {
	now := utils.NowMillis()
	return helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		fillModel := models.NewTTradeFillModel(conn, l.svcCtx.Config.CacheRedis)
		reservationModel := models.NewTTradeAssetReservationModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)

		current, err := instructionModel.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
		if !settlementInstructionLeaseOwned(current, item) {
			return fmt.Errorf("settlement instruction is not processing")
		}
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)
		current.LastErrorMsg = ""
		current.NextRetryAt = 0
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}

		reservation, err := reservationModel.FindOneByReservationNoForUpdate(ctx, item.TenantId, item.ReservationNo)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if reservation != nil {
			var ok bool
			switch trade.SettlementInstructionAction(item.Action) {
			case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN:
				ok, err = reservationModel.AddConsumed(ctx, reservation.Id, item.Amount, now)
			case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE:
				if fill.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) || order.Side == int64(common.Side_SIDE_BUY) {
					ok, err = reservationModel.AddConsumed(ctx, reservation.Id, item.Amount, now)
				} else {
					ok = true
				}
			case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN:
				ok, err = reservationModel.AddReleased(ctx, reservation.Id, item.Amount, now)
			case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_ADJUST_MARGIN:
				ok, err = reservationModel.AddConsumed(ctx, reservation.Id, item.Amount, now)
			default:
				ok = true
			}
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("reservation settlement amount exceeds reserved amount")
			}
		}

		if err := ensureFillRemainderRelease(ctx, instructionModel, reservationModel, order, fill, now); err != nil {
			return err
		}
		instructions, err := instructionModel.FindByFillId(ctx, fill.TenantId, fill.Id)
		if err != nil {
			return err
		}
		for _, instruction := range instructions {
			if instruction.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
				return nil
			}
		}
		currentFill, err := fillModel.FindOne(ctx, fill.Id)
		if err != nil {
			return err
		}
		if currentFill.SettlementStatus == int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_SETTLED) {
			return nil
		}
		if fill.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
			positionEvent, err := eventModel.FindOneByTenantIdEventNo(ctx, fill.TenantId, derivedTradeBizNo(fill.FillNo, "POSITION"))
			if err != nil {
				return err
			}
			if positionEvent.EventStatus != int64(trade.EventStatus_EVENT_STATUS_SUCCESS) {
				return nil
			}
		}
		currentFill.SettlementStatus = int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_SETTLED)
		currentFill.SettledAt = now
		if err := fillModel.Update(ctx, currentFill); err != nil {
			return err
		}
		eventType := "SPOT_FILL_SETTLED"
		if fill.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
			eventType = "CONTRACT_FILL_ASSET_SETTLED"
		}
		if err := insertMatchOutboxEvent(ctx, eventModel, order, derivedTradeBizNo(fill.FillNo, "SETTLED"), eventType, fill.FillNo, "fill", "{}", now); err != nil {
			return err
		}
		if _, err := finalizeOrderTermination(ctx, conn, l.svcCtx, order.Id, now); err != nil {
			return err
		}
		return finalizeSettledOrder(ctx, conn, l.svcCtx, order.Id, now)
	})
}

func (l *ProcessFillSettlementsLogic) settleFillIfReady(fill *models.TTradeFill) error {
	if fill == nil || fill.SettlementStatus == int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_SETTLED) {
		return nil
	}
	now := utils.NowMillis()
	return helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		fillModel := models.NewTTradeFillModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)
		instructions, err := instructionModel.FindByFillId(ctx, fill.TenantId, fill.Id)
		if err != nil {
			return err
		}
		for _, instruction := range instructions {
			if instruction.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
				return nil
			}
		}
		if fill.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
			positionEvent, err := eventModel.FindOneByTenantIdEventNo(ctx, fill.TenantId, derivedTradeBizNo(fill.FillNo, "POSITION"))
			if err != nil {
				return err
			}
			if positionEvent.EventStatus != int64(trade.EventStatus_EVENT_STATUS_SUCCESS) {
				return nil
			}
		}
		current, err := fillModel.FindOne(ctx, fill.Id)
		if err != nil {
			return err
		}
		if current.SettlementStatus == int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_SETTLED) {
			return nil
		}
		current.SettlementStatus = int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_SETTLED)
		current.SettledAt = now
		if err := fillModel.Update(ctx, current); err != nil {
			return err
		}
		order, err := orderModel.FindOne(ctx, fill.OrderId)
		if err != nil {
			return err
		}
		eventType := "SPOT_FILL_SETTLED"
		if fill.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
			eventType = "CONTRACT_FILL_ASSET_SETTLED"
		}
		if err := insertMatchOutboxEvent(ctx, eventModel, order, derivedTradeBizNo(fill.FillNo, "SETTLED"), eventType, fill.FillNo, "fill", "{}", now); err != nil {
			return err
		}
		if _, err := finalizeOrderTermination(ctx, conn, l.svcCtx, order.Id, now); err != nil {
			return err
		}
		return finalizeSettledOrder(ctx, conn, l.svcCtx, order.Id, now)
	})
}

func ensureFillRemainderRelease(ctx context.Context, instructionModel models.TTradeSettlementInstructionModel, reservationModel models.TTradeAssetReservationModel, order *models.TTradeOrder, fill *models.TTradeFill, now int64) error {
	if order.Status != int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING) && order.Status != int64(trade.OrderStatus_ORDER_STATUS_CANCELING) && order.Status != int64(trade.OrderStatus_ORDER_STATUS_EXPIRING) {
		return nil
	}
	unfinished, err := instructionModel.CountUnfinishedByOrder(ctx, order.TenantId, order.Id)
	if err != nil || unfinished > 0 {
		return err
	}
	reservation, err := reservationModel.FindOneByReservationNoForUpdate(ctx, order.TenantId, order.OrderNo)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	remaining := reservation.ReservedAmount.Sub(reservation.ConsumedAmount).Sub(reservation.ReleasedAmount)
	if !remaining.IsPositive() {
		return nil
	}
	if _, err := reservationModel.BeginRelease(ctx, reservation.Id, now); err != nil {
		return err
	}
	// The remainder belongs to the order reservation, not to an individual Fill.
	// A stable order-level key prevents concurrent final Fills from creating two
	// independently idempotent unfreeze requests for the same balance.
	return insertSettlementInstructionIdempotent(ctx, instructionModel, &models.TTradeSettlementInstruction{TenantId: fill.TenantId, InstructionNo: derivedTradeBizNo(order.OrderNo, "RELEASE"), BizType: "order", BizId: order.OrderNo, OrderId: order.Id, ReservationNo: order.OrderNo, UserId: order.UserId, Action: int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN), Asset: reservation.Asset, Amount: remaining, StepNo: 1, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now})
}

func (l *ProcessFillSettlementsLogic) markFailed(item *models.TTradeSettlementInstruction, cause error) error {
	now := utils.NowMillis()
	return helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		fillModel := models.NewTTradeFillModel(conn, l.svcCtx.Config.CacheRedis)
		reservationModel := models.NewTTradeAssetReservationModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := instructionModel.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) || !settlementInstructionLeaseOwned(current, item) {
			return nil
		}
		current.RetryCount++
		current.Status, current.NextRetryAt = settlementFailureTransition(current.RetryCount, now)
		current.LastErrorMsg = cause.Error()
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}
		fill, err := fillModel.FindOne(ctx, item.FillId)
		if err != nil {
			return err
		}
		fill.SettlementStatus = int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_FAILED)
		fill.SettlementRetryCount++
		if err := fillModel.Update(ctx, fill); err != nil {
			return err
		}
		action := trade.SettlementInstructionAction(item.Action)
		tracksReservation := action == trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN ||
			action == trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN ||
			action == trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_ADJUST_MARGIN ||
			action == trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE && (fill.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) || fill.Side == int64(common.Side_SIDE_BUY))
		if !tracksReservation {
			return nil
		}
		reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, item.TenantId, item.ReservationNo)
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		retryStatus := reservation.Status
		if action == trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN {
			retryStatus = int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_RELEASING)
		}
		terminal := current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW)
		return reservationModel.MarkSettlementFailure(ctx, reservation.Id, retryStatus, terminal, current.NextRetryAt, cause.Error(), now)
	})
}

func settlementFailureTransition(retryCount, now int64) (status, nextRetryAt int64) {
	if retryCount >= spotSettlementMaxRetry {
		return int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW), 0
	}
	delaySeconds := int64(1) << min(retryCount, int64(10))
	return int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED), now + delaySeconds*1000
}
