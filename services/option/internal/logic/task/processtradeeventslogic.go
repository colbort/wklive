package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/observability"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

// 按合约撮合序号消费成交持仓事件
func (l *ProcessTradeEventsLogic) ProcessTradeEvents(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_trade_events", func() (*option.OptionTaskResp, error) {
		now := time.Now().Unix()
		if err := l.svcCtx.OptionOutboxModel.RecoverStale(l.ctx, now-60, now); err != nil {
			return nil, err
		}
		staleBarrierEvents, metricErr := l.svcCtx.OptionOutboxModel.
			CountStaleComboDebitBarrierBlocked(l.ctx, in.TenantId, now-60)
		if metricErr != nil {
			observability.RecordComboObservabilityQueryFailure(
				in.TenantId, "stale_debit_barrier",
			)
			l.Errorf("query stale option combo debit barriers failed: %v", metricErr)
		} else {
			observability.SetComboDebitBarrierStaleEvents(in.TenantId, staleBarrierEvents)
		}
		if sampleErr := observability.SampleOptionOperationsMetrics(
			l.ctx, l.svcCtx.DB, now,
		); sampleErr != nil {
			l.Errorf("sample option operations metrics failed: %v", sampleErr)
		}
		for {
			events, err := l.svcCtx.OptionOutboxModel.FindRunnable(l.ctx, in.TenantId, now, 100)
			if err != nil {
				return nil, err
			}
			if len(events) == 0 {
				return helpers.OkTaskResp(), nil
			}
			for _, event := range events {
				if err := l.processOneTradeEvent(event); err != nil {
					l.Errorf("process option trade event failed, eventNo=%s err=%v", event.EventNo, err)
				}
			}
			if len(events) < 100 {
				return helpers.OkTaskResp(), nil
			}
			now = time.Now().Unix()
		}
	})
}

func (l *ProcessTradeEventsLogic) processOneTradeEvent(event *models.TOptionOutbox) error {
	now := time.Now().Unix()
	claimed, err := l.svcCtx.OptionOutboxModel.Claim(l.ctx, event.Id, now)
	if err != nil || !claimed {
		return err
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		outboxModel := models.NewTOptionOutboxModel(conn, l.svcCtx.Config.CacheRedis)
		inboxModel := models.NewTOptionInboxModel(conn, l.svcCtx.Config.CacheRedis)
		tradeModel := models.NewTOptionTradeModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)

		current, err := outboxModel.FindOneForUpdate(ctx, event.Id)
		if err != nil {
			return err
		}
		if current.Status != int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PROCESSING) {
			return nil
		}
		if current.EventType != int64(option.OptionEventType_OPTION_EVENT_TYPE_TRADE_POSITION) {
			return fmt.Errorf("unsupported option event type: %d", current.EventType)
		}
		inbox, err := inboxModel.FindOneByTenantIdEventNo(ctx, current.TenantId, current.EventNo)
		if err == nil && inbox.Status == int64(option.OptionInboxStatus_OPTION_INBOX_STATUS_SUCCESS) {
			current.Status = int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS)
			current.LastErrorMsg = ""
			current.NextRetryAt = 0
			current.UpdateTimes = now
			return outboxModel.Update(ctx, current)
		}
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if errors.Is(err, models.ErrNotFound) {
			inbox = &models.TOptionInbox{
				TenantId: current.TenantId, EventNo: current.EventNo, EventType: current.EventType,
				ContractId: current.ContractId, TradeId: current.TradeId, MatchSequence: current.MatchSequence,
				Status:      int64(option.OptionInboxStatus_OPTION_INBOX_STATUS_PROCESSING),
				CreateTimes: now, UpdateTimes: now,
			}
			result, err := inboxModel.Insert(ctx, inbox)
			if err != nil {
				return err
			}
			inbox.Id, err = result.LastInsertId()
			if err != nil {
				return err
			}
		}

		trade, err := tradeModel.FindOne(ctx, current.TradeId)
		if err != nil {
			return err
		}
		if trade.TenantId != current.TenantId || trade.ContractId != current.ContractId ||
			trade.MatchSequence != current.MatchSequence {
			return errors.New("option trade event does not match trade")
		}
		if trade.ComboMatchNo != "" {
			barrierReady, barrierErr := outboxModel.ComboDebitBarrierReady(
				ctx, trade.TenantId, trade.ComboMatchNo,
			)
			if barrierErr != nil {
				return barrierErr
			}
			if !barrierReady {
				observability.RecordComboDebitBarrierViolation(trade.TenantId)
				return fmt.Errorf(
					"option combo debit barrier is incomplete: %s", trade.ComboMatchNo,
				)
			}
		}
		contract, err := contractModel.FindOne(ctx, trade.ContractId)
		if err != nil {
			return err
		}
		buyOrder, err := orderModel.FindOne(ctx, trade.BuyOrderId)
		if err != nil {
			return err
		}
		sellOrder, err := orderModel.FindOne(ctx, trade.SellOrderId)
		if err != nil {
			return err
		}
		if err := updatePositionByFilledOrder(
			ctx, positionModel, contract, buyOrder,
			trade.Price, trade.Qty, trade.BuyFee, trade.TradeTime,
		); err != nil {
			return err
		}
		if err := updatePositionByFilledOrder(
			ctx, positionModel, contract, sellOrder,
			trade.Price, trade.Qty, trade.SellFee, trade.TradeTime,
		); err != nil {
			return err
		}
		if buyOrder.Side == int64(common.Side_SIDE_BUY) &&
			buyOrder.PositionEffect == int64(option.PositionEffect_POSITION_EFFECT_CLOSE) {
			position, err := positionModel.FindOneByTenantIdUserIdAccountIdContractIdSide(
				ctx, buyOrder.TenantId, buyOrder.UserId, buyOrder.AccountId, buyOrder.ContractId,
				int64(common.PositionSide_POSITION_SIDE_SHORT),
			)
			if err != nil {
				return err
			}
			if err := createCloseMarginReleaseInstructions(
				ctx, marginLotModel, instructionModel, contract, position, trade, buyOrder, now,
			); err != nil {
				return err
			}
		}
		if sellOrder.Side == int64(common.Side_SIDE_SELL) &&
			sellOrder.PositionEffect == int64(option.PositionEffect_POSITION_EFFECT_OPEN) &&
			(contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) ||
				contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY)) {
			lot, err := marginLotModel.FindOneByTenantIdTradeId(ctx, trade.TenantId, trade.Id)
			if err != nil {
				return err
			}
			position, err := positionModel.FindOneByTenantIdUserIdAccountIdContractIdSide(
				ctx, sellOrder.TenantId, sellOrder.UserId, sellOrder.AccountId, sellOrder.ContractId,
				int64(common.PositionSide_POSITION_SIDE_SHORT),
			)
			if err != nil {
				return err
			}
			position.MarginAmount = position.MarginAmount.Add(lot.InitialMargin)
			position.UpdateTimes = now
			if err := positionModel.Update(ctx, position); err != nil {
				return err
			}
			lot.PositionId = position.Id
			lot.UpdateTimes = now
			if err := marginLotModel.Update(ctx, lot); err != nil {
				return err
			}
		}
		inbox.Status = int64(option.OptionInboxStatus_OPTION_INBOX_STATUS_SUCCESS)
		inbox.LastErrorMsg = ""
		inbox.UpdateTimes = now
		if err := inboxModel.Update(ctx, inbox); err != nil {
			return err
		}
		current.Status = int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS)
		current.LastErrorMsg = ""
		current.NextRetryAt = 0
		current.UpdateTimes = now
		return outboxModel.Update(ctx, current)
	})
	if err != nil {
		return errors.Join(err, l.markTradeEventFailed(event, err))
	}
	return nil
}

func createCloseMarginReleaseInstructions(
	ctx context.Context,
	marginLotModel models.TOptionMarginLotModel,
	instructionModel models.TOptionAssetInstructionModel,
	contract *models.TOptionContract,
	position *models.TOptionPosition,
	trade *models.TOptionTrade,
	order *models.TOptionOrder,
	now int64,
) error {
	lots, err := marginLotModel.FindClosableByPosition(ctx, trade.TenantId, position.Id)
	if err != nil {
		return err
	}
	remainingQty := trade.Qty
	for _, lot := range lots {
		if !remainingQty.IsPositive() {
			break
		}
		closeQty := decimal.Min(remainingQty, lot.RemainingQuantity)
		if !closeQty.IsPositive() {
			continue
		}
		availableMargin := decimal.Max(lot.RemainingMargin.Sub(lot.PendingMargin), decimal.Zero)
		releaseAmount := availableMargin
		if closeQty.LessThan(lot.RemainingQuantity) {
			releaseAmount = availableMargin.Mul(closeQty).Div(lot.RemainingQuantity).Round(16)
		}
		if releaseAmount.IsPositive() {
			collateralCoin := lot.CollateralCoin
			if collateralCoin == "" {
				return fmt.Errorf("option margin lot collateral coin evidence is missing: lotId=%d", lot.Id)
			}
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: trade.TenantId, InstructionNo: fmt.Sprintf("%s-MARGIN-%d", trade.TradeNo, lot.Id),
				BizNo: trade.TradeNo, OrderId: order.Id, TradeId: trade.Id,
				PositionId: position.Id, MarginLotId: lot.Id,
				UserId: order.UserId, AccountId: order.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: lot.FreezeBizNo, Coin: collateralCoin, Amount: releaseAmount,
				StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return err
			}
			lot.PendingMargin = lot.PendingMargin.Add(releaseAmount)
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
		}
		lot.RemainingQuantity = decimal.Max(lot.RemainingQuantity.Sub(closeQty), decimal.Zero)
		lot.UpdateTimes = now
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return err
		}
		remainingQty = remainingQty.Sub(closeQty)
	}
	if remainingQty.IsPositive() {
		return fmt.Errorf("insufficient seller margin lots for close quantity: %s", remainingQty.String())
	}
	return nil
}

func (l *ProcessTradeEventsLogic) markTradeEventFailed(event *models.TOptionOutbox, cause error) error {
	current, err := l.svcCtx.OptionOutboxModel.FindOne(l.ctx, event.Id)
	if err != nil {
		return err
	}
	current.RetryCount++
	current.Status = int64(option.OptionEventStatus_OPTION_EVENT_STATUS_FAILED)
	if current.RetryCount >= 20 {
		current.Status = int64(option.OptionEventStatus_OPTION_EVENT_STATUS_MANUAL_REVIEW)
	}
	current.NextRetryAt = time.Now().Add(optionAssetRetryDelay(current.RetryCount)).Unix()
	current.LastErrorMsg = cause.Error()
	if len(current.LastErrorMsg) > 500 {
		current.LastErrorMsg = current.LastErrorMsg[:500]
	}
	current.UpdateTimes = time.Now().Unix()
	return l.svcCtx.OptionOutboxModel.Update(l.ctx, current)
}
