package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const reservationReleaseBatchSize = int64(100)

type ProcessReservationReleasesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessReservationReleasesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessReservationReleasesLogic {
	return &ProcessReservationReleasesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ProcessReservationReleasesLogic) Process(tenantID int64) error {
	if err := l.recoverTerminatingOrders(tenantID); err != nil {
		return err
	}
	for processed := 0; processed < spotSettlementMaxSteps; {
		now := utils.NowMillis()
		items, err := l.svcCtx.TradeSettlementInstrModel.FindPendingOrderReleases(l.ctx, tenantID, now, reservationReleaseBatchSize)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		progressed := false
		for _, item := range items {
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
			if err := l.executeClaimed(item); err != nil {
				l.Errorf("reservation release failed, instructionNo=%s orderId=%d err=%v", item.InstructionNo, item.OrderId, err)
				if markErr := l.markFailed(item, err); markErr != nil {
					return markErr
				}
			}
		}
		if !progressed {
			return nil
		}
	}
	return nil
}

// recoverTerminatingOrders closes the crash window between persisting an
// order's CANCELING/EXPIRING state and creating its durable release
// instruction. It is intentionally idempotent: the instruction number and
// reservation transition are protected by unique keys and row locks.
func (l *ProcessReservationReleasesLogic) recoverTerminatingOrders(tenantID int64) error {
	cursor := int64(0)
	for scanned := 0; scanned < spotSettlementMaxSteps; {
		orders, _, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
			TenantId: tenantID,
			Statuses: helpers.TerminatingOrderStatuses(),
		}, cursor, reservationReleaseBatchSize)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, order := range orders {
			cursor = order.Id
			scanned++
			if err = unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, order, "recover terminating order reservation release"); err != nil {
				return err
			}
			if err = removeOrderBookOrder(l.svcCtx, l.ctx, order); err != nil {
				l.Errorf("repair terminating order cache failed, orderId=%d err=%v", order.Id, err)
			}
			if scanned >= spotSettlementMaxSteps {
				return nil
			}
		}
		if int64(len(orders)) < reservationReleaseBatchSize {
			return nil
		}
	}
	return nil
}

func (l *ProcessReservationReleasesLogic) ProcessInstruction(instructionID int64) error {
	item, err := l.svcCtx.TradeSettlementInstrModel.FindOne(l.ctx, instructionID)
	if err != nil {
		return err
	}
	if item.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
		return nil
	}
	claimed, lease, err := l.svcCtx.TradeSettlementInstrModel.ClaimLease(l.ctx, item.Id, utils.NowMillis())
	if err != nil || !claimed {
		return err
	}
	item.UpdateTimes = lease
	if err := l.executeClaimed(item); err != nil {
		if markErr := l.markFailed(item, err); markErr != nil {
			return markErr
		}
		return err
	}
	return nil
}

func (l *ProcessReservationReleasesLogic) executeClaimed(item *models.TTradeSettlementInstruction) error {
	if item.BizType != "order" || item.Action != int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN) || item.ReservationNo == "" || !item.Amount.IsPositive() {
		return fmt.Errorf("invalid order reservation release instruction")
	}
	// Execute the immutable instruction amount. Asset idempotency is keyed by
	// BizNo, so a retry after Asset committed but before Trade checkpointed is
	// still safe, while the generated flow remains exactly reconcilable.
	resp, err := l.svcCtx.AssetClient.UnfreezeAssetByBizNo(l.ctx, &asset.UnfreezeAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_CANCEL_ORDER, BizId: item.OrderId, BizNo: item.InstructionNo, Remark: "trade order reservation release"})
	if err != nil {
		if i18n.IsStatusError(err, i18n.FreezeRecordNotFound) {
			return l.markMissingFreezeSucceeded(item)
		}
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(l.ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(l.ctx, resp.Base.Code)
	}
	return l.markSucceeded(item)
}

// markMissingFreezeSucceeded handles a definitive Asset response proving that
// the original freeze never committed. There is nothing to unfreeze, but the
// local reservation and release instruction must be closed so they do not
// block all later recovery work.
func (l *ProcessReservationReleasesLogic) markMissingFreezeSucceeded(item *models.TTradeSettlementInstruction) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		instructionModel := tx.TradeSettlementInstruction
		reservationModel := tx.TradeAssetReservation
		current, err := instructionModel.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
		if !settlementInstructionLeaseOwned(current, item) {
			return fmt.Errorf("reservation release instruction is not processing")
		}
		reservation, err := reservationModel.FindOneByReservationNoForUpdate(ctx, item.TenantId, item.ReservationNo)
		if err != nil {
			return err
		}
		if reservation.ConsumedAmount.IsPositive() {
			return fmt.Errorf("asset freeze is missing after reservation consumption")
		}
		reservation.Status = int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_FAILED)
		reservation.ReleasedAmount = reservation.ReservedAmount
		reservation.NextRetryAt = 0
		reservation.LastErrorMsg = "asset freeze record not found; no funds were frozen"
		reservation.UpdateTimes = now
		reservation.Version++
		if err := reservationModel.Update(ctx, reservation); err != nil {
			return err
		}
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)
		current.NextRetryAt = 0
		current.LastErrorMsg = "asset freeze record not found; treated as no-op release"
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}
		_, err = finalizeOrderTermination(ctx, tx, item.OrderId, now)
		return err
	})
}

func (l *ProcessReservationReleasesLogic) markSucceeded(item *models.TTradeSettlementInstruction) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		instructionModel := tx.TradeSettlementInstruction
		reservationModel := tx.TradeAssetReservation
		current, err := instructionModel.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
		if !settlementInstructionLeaseOwned(current, item) {
			return fmt.Errorf("reservation release instruction is not processing")
		}
		reservation, err := reservationModel.FindOneByReservationNoForUpdate(ctx, item.TenantId, item.ReservationNo)
		if err != nil {
			return err
		}
		ok, err := reservationModel.AddReleased(ctx, reservation.Id, current.Amount, now)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("reservation release exceeds remaining amount")
		}
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)
		current.NextRetryAt = 0
		current.LastErrorMsg = ""
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}
		_, err = finalizeOrderTermination(ctx, tx, item.OrderId, now)
		return err
	})
}

func (l *ProcessReservationReleasesLogic) markFailed(item *models.TTradeSettlementInstruction, cause error) error {
	now := utils.NowMillis()
	return l.svcCtx.TransactionModel.Transact(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		instructionModel := tx.TradeSettlementInstruction
		reservationModel := tx.TradeAssetReservation
		current, err := instructionModel.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) || !settlementInstructionLeaseOwned(current, item) {
			return nil
		}
		current.RetryCount++
		terminal := current.RetryCount >= spotSettlementMaxRetry
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED)
		current.NextRetryAt = now + (int64(1)<<min(current.RetryCount, int64(10)))*1000
		if terminal {
			current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW)
			current.NextRetryAt = 0
		}
		current.LastErrorMsg = cause.Error()
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}
		reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, item.TenantId, item.ReservationNo)
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		return reservationModel.MarkSettlementFailure(ctx, reservation.Id, int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_RELEASING), terminal, current.NextRetryAt, cause.Error(), now)
	})
}
