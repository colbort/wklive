package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wklive/services/option/internal/logic/helpers"

	"wklive/common/generate"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessContractLifecycleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessContractLifecycleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessContractLifecycleLogic {
	return &ProcessContractLifecycleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 期权合约生命周期处理（状态流转/订单过期/自动行权/到期结算）
func (l *ProcessContractLifecycleLogic) ProcessContractLifecycle(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_contract_lifecycle", func() (*option.OptionTaskResp, error) {
		now := time.Now().Unix()
		if err := l.syncContracts(option.ContractStatus_CONTRACT_STATUS_PENDING, now, 0, option.ContractStatus_CONTRACT_STATUS_TRADING, now); err != nil {
			return nil, err
		}
		if err := l.syncContracts(option.ContractStatus_CONTRACT_STATUS_TRADING, 0, now, option.ContractStatus_CONTRACT_STATUS_EXPIRED, now); err != nil {
			return nil, err
		}
		if err := l.processExpiredContracts(now); err != nil {
			return nil, err
		}
		return helpers.OkTaskResp(), nil
	})
}

func (l *ProcessContractLifecycleLogic) syncContracts(from option.ContractStatus, listEnd, expireEnd int64, to option.ContractStatus, now int64) error {
	cursor := int64(0)
	for {
		items, _, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
			Status:        int64(from),
			ListTimeEnd:   listEnd,
			ExpireTimeEnd: expireEnd,
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for _, item := range items {
			cursor = item.Id
			item.Status = int64(to)
			item.UpdateTimes = now
			if err := l.svcCtx.OptionContractModel.Update(l.ctx, item); err != nil {
				return err
			}
		}
		if len(items) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) processExpiredContracts(now int64) error {
	cursor := int64(0)
	for {
		contracts, _, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
			Status: int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(contracts) == 0 {
			return nil
		}
		for _, contract := range contracts {
			cursor = contract.Id
			if err := l.expireContractOrders(contract, now); err != nil {
				return err
			}
			incomplete, err := l.svcCtx.OptionOutboxModel.HasIncomplete(l.ctx, contract.TenantId, contract.Id)
			if err != nil {
				return err
			}
			if incomplete {
				l.Infof("delay option settlement for incomplete trade events, contractId=%d", contract.Id)
				continue
			}
			incomplete, err = l.svcCtx.OptionAssetInstructionModel.HasIncompleteMarginForContract(
				l.ctx, contract.TenantId, contract.Id,
			)
			if err != nil {
				return err
			}
			if incomplete {
				l.Infof("delay option settlement for incomplete margin instructions, contractId=%d", contract.Id)
				continue
			}
			pendingExercise, err := l.svcCtx.OptionExerciseModel.HasPendingByContract(
				l.ctx, contract.TenantId, contract.Id,
			)
			if err != nil {
				return err
			}
			if pendingExercise {
				l.Infof("delay option settlement for pending early exercises, contractId=%d", contract.Id)
				continue
			}
			settlementPrice, err := l.lockSettlementPrice(contract, now)
			if err != nil {
				return err
			}
			if settlementPrice == nil ||
				settlementPrice.Status != int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED) {
				l.Errorf("option settlement waiting for authoritative price, contractId=%d", contract.Id)
				continue
			}
			if contract.IsAutoExercise == int64(common.YesNo_YES_NO_YES) {
				if err := l.autoExerciseContract(contract, settlementPrice.DeliveryPrice, now); err != nil {
					return err
				}
			}
			if contract.DeliverTime > now {
				continue
			}
			if err := l.settleContract(contract, settlementPrice, now); err != nil {
				return err
			}
		}
		if len(contracts) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) expireContractOrders(contract *models.TOptionContract, now int64) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.OptionOrderModel.FindPage(l.ctx, models.OptionOrderPageFilter{
			ContractId: contract.Id,
			Statuses: []int64{
				int64(option.OrderStatus_ORDER_STATUS_PENDING),
				int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
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
			order.Status = int64(option.OrderStatus_ORDER_STATUS_EXPIRED)
			if order.MarginAmount.IsPositive() {
				order.Status = int64(option.OrderStatus_ORDER_STATUS_EXPIRING)
			}
			order.CancelReason = "CONTRACT_EXPIRED"
			order.CancelTime = now
			order.UpdateTimes = now
			err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				conn := sqlx.NewSqlConnFromSession(session)
				orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
				positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
				instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
				if err := releaseClosePositionFrozenQty(ctx, positionModel, order, order.UnfilledQty, now); err != nil {
					return err
				}
				if order.MarginAmount.IsPositive() {
					if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
						TenantId: order.TenantId, InstructionNo: order.OrderNo + "-EXPIRE-RELEASE",
						BizNo: order.OrderNo, OrderId: order.Id, UserId: order.UserId, AccountId: order.AccountId,
						Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
						TargetBizNo: order.OrderNo, Coin: order.FeeCoin, Amount: order.MarginAmount,
						StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
						ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
						CreateTimes:          now, UpdateTimes: now,
					}); err != nil {
						return err
					}
				}
				return orderModel.Update(ctx, order)
			})
			if err != nil {
				return err
			}
			publishOptionOrderChanged(l.ctx, l.svcCtx, order)
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) lockSettlementPrice(contract *models.TOptionContract, now int64) (*models.TOptionSettlementPrice, error) {
	item, err := l.svcCtx.OptionSettlementPriceModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	const maxSettlementQuoteAge = int64(300)
	if !market.UnderlyingPrice.IsPositive() ||
		market.SnapshotTime <= 0 ||
		market.SnapshotTime > now ||
		now-market.SnapshotTime > maxSettlementQuoteAge {
		return nil, nil
	}
	item = &models.TOptionSettlementPrice{
		TenantId: contract.TenantId, ContractId: contract.Id,
		PriceSource: "authoritative-market", WindowStart: market.SnapshotTime,
		WindowEnd: market.SnapshotTime, SampleCount: 1, CalculationMethod: "LAST_AT_EXPIRY",
		DeliveryPrice:     market.UnderlyingPrice,
		SourceSnapshotIds: fmt.Sprintf("market:%d:%d", market.Id, market.SnapshotTime),
		Version:           1, Status: int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED),
		ConfirmedAt: now, CreateTimes: now, UpdateTimes: now,
	}
	result, err := l.svcCtx.OptionSettlementPriceModel.Insert(l.ctx, item)
	if err != nil {
		existing, findErr := l.svcCtx.OptionSettlementPriceModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
		if findErr == nil {
			return existing, nil
		}
		return nil, err
	}
	item.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (l *ProcessContractLifecycleLogic) autoExerciseContract(contract *models.TOptionContract, deliveryPrice decimal.Decimal, now int64) error {
	intrinsicValue := optionIntrinsicValue(contract, deliveryPrice)
	cursor := int64(0)
	for {
		positions, _, err := l.svcCtx.OptionPositionModel.FindPage(l.ctx, models.OptionPositionPageFilter{
			ContractId: contract.Id,
			Status:     int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(positions) == 0 {
			return nil
		}
		for _, position := range positions {
			cursor = position.Id
			if err := l.autoExercisePosition(contract, position.Id, deliveryPrice, intrinsicValue, now); err != nil {
				return err
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) autoExercisePosition(contract *models.TOptionContract, positionId int64, deliveryPrice, intrinsicValue decimal.Decimal, now int64) error {
	// Generate outside the database transaction so a Redis/network round-trip
	// cannot hold the position row lock.
	exerciseNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "EX", "")
	if err != nil {
		return err
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		exerciseModel := models.NewTOptionExerciseModel(conn, l.svcCtx.Config.CacheRedis)

		position, err := positionModel.FindOneForUpdate(ctx, positionId)
		if err != nil {
			return err
		}
		if position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) {
			return nil
		}
		if position.Side != int64(common.PositionSide_POSITION_SIDE_LONG) ||
			!position.ExerciseableQty.IsPositive() || !intrinsicValue.IsPositive() {
			position.Status = int64(option.PositionStatus_POSITION_STATUS_EXPIRED)
			position.ExerciseableQty = decimal.Zero
			position.UpdateTimes = now
			return positionModel.Update(ctx, position)
		}

		exists, _, err := exerciseModel.FindPage(ctx, models.OptionExercisePageFilter{
			TenantId: position.TenantId, PositionId: position.Id,
			ExerciseType: int64(option.ExerciseType_EXERCISE_TYPE_AUTO),
		}, 0, 1)
		if err != nil {
			return err
		}
		if len(exists) > 0 {
			return nil
		}
		if _, err = exerciseModel.Insert(ctx, &models.TOptionExercise{
			TenantId: position.TenantId, ExerciseNo: exerciseNo, UserId: position.UserId,
			AccountId: position.AccountId, ContractId: contract.Id, PositionId: position.Id,
			ExerciseType: int64(option.ExerciseType_EXERCISE_TYPE_AUTO), ExerciseQty: position.ExerciseableQty,
			StrikePrice: contract.StrikePrice, SettlementPrice: deliveryPrice,
			ExerciseAmount: optionExerciseAmount(contract, position.ExerciseableQty),
			ProfitAmount:   optionSettlementPayoff(contract, deliveryPrice, position.ExerciseableQty),
			Fee:            optionSettlementPayoff(contract, deliveryPrice, position.ExerciseableQty).Mul(contract.ExerciseFeeRate).Round(16),
			FeeCoin:        contract.SettleCoin, Status: int64(option.ExerciseStatus_EXERCISE_STATUS_DONE),
			Remark: "option auto exercise task", ExerciseTime: now, FinishTime: now,
			CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		position.Status = int64(option.PositionStatus_POSITION_STATUS_EXERCISED)
		position.ExerciseableQty = decimal.Zero
		position.UpdateTimes = now
		return positionModel.Update(ctx, position)
	})
}

func (l *ProcessContractLifecycleLogic) settleContract(contract *models.TOptionContract, settlementPrice *models.TOptionSettlementPrice, now int64) error {
	_, err := l.svcCtx.OptionSettlementModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	deliveryPrice := settlementPrice.DeliveryPrice
	theoreticalPrice := decimal.Zero
	iv := decimal.Zero
	isITM := int64(common.YesNo_YES_NO_NO)
	if market != nil {
		theoreticalPrice = market.TheoreticalPrice
		iv = market.Iv
		if (contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) && deliveryPrice.GreaterThan(contract.StrikePrice)) ||
			(contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) && deliveryPrice.LessThan(contract.StrikePrice)) {
			isITM = int64(common.YesNo_YES_NO_YES)
		}
	}
	exerciseResult := int64(option.ExerciseResult_EXERCISE_RESULT_NONE)
	if contract.IsAutoExercise == int64(common.YesNo_YES_NO_YES) {
		if isITM == int64(common.YesNo_YES_NO_YES) {
			exerciseResult = int64(option.ExerciseResult_EXERCISE_RESULT_AUTO_EXERCISE)
		} else {
			exerciseResult = int64(option.ExerciseResult_EXERCISE_RESULT_AUTO_ABANDON)
		}
	}
	settlementNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "OPS", "")
	if err != nil {
		return err
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		settlementModel := models.NewTOptionSettlementModel(conn, l.svcCtx.Config.CacheRedis)
		batchModel := models.NewTOptionSettlementBatchModel(conn, l.svcCtx.Config.CacheRedis)
		detailModel := models.NewTOptionSettlementDetailModel(conn, l.svcCtx.Config.CacheRedis)
		snapshotModel := models.NewTOptionMarketSnapshotModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)

		batchResult, err := batchModel.Insert(ctx, &models.TOptionSettlementBatch{
			TenantId: contract.TenantId, BatchNo: settlementNo, ContractId: contract.Id,
			SettlementPriceId: settlementPrice.Id,
			Status:            int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_CALCULATING),
			CreateTimes:       now, UpdateTimes: now,
		})
		if err != nil {
			return err
		}
		batchId, err := batchResult.LastInsertId()
		if err != nil {
			return err
		}
		result, err := settlementModel.Insert(ctx, &models.TOptionSettlement{
			TenantId:         contract.TenantId,
			SettlementNo:     settlementNo,
			ContractId:       contract.Id,
			UnderlyingSymbol: contract.UnderlyingSymbol,
			ExpireTime:       contract.ExpireTime,
			SettlementTime:   now,
			DeliveryPrice:    deliveryPrice,
			TheoreticalPrice: theoreticalPrice,
			Iv:               iv,
			IsItm:            isITM,
			ExerciseResult:   exerciseResult,
			Status:           int64(option.SettlementStatus_SETTLEMENT_STATUS_PROCESSING),
			Remark:           "option settlement task",
			CreateTimes:      now,
			UpdateTimes:      now,
		})
		if err != nil {
			return err
		}
		settlementId, err := result.LastInsertId()
		if err != nil {
			return err
		}
		summary, err := settleContractPositions(ctx, positionModel, detailModel, instructionModel, marginLotModel, contract, batchId, settlementNo, deliveryPrice, now)
		if err != nil {
			return err
		}
		batch, err := batchModel.FindOne(ctx, batchId)
		if err != nil {
			return err
		}
		batch.TotalCredit = summary.totalCredit
		batch.TotalDebit = summary.totalDebit
		batch.InstructionCount = summary.instructionCount
		batch.Status = int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_INSTRUCTIONS_CREATED)
		if summary.instructionCount == 0 {
			batch.Status = int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_DONE)
		}
		batch.UpdateTimes = now
		if err := batchModel.Update(ctx, batch); err != nil {
			return err
		}
		if err := helpers.InsertMarketSnapshot(ctx, snapshotModel, market, now); err != nil {
			return err
		}
		if summary.instructionCount == 0 {
			settlement, err := settlementModel.FindOne(ctx, settlementId)
			if err != nil {
				return err
			}
			settlement.Status = int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE)
			settlement.UpdateTimes = now
			if err := settlementModel.Update(ctx, settlement); err != nil {
				return err
			}
			contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_SETTLED)
			contract.UpdateTimes = now
			return contractModel.Update(ctx, contract)
		}
		return nil
	})
}

type optionSettlementSummary struct {
	totalCredit      decimal.Decimal
	totalDebit       decimal.Decimal
	instructionCount int64
}

func settleContractPositions(ctx context.Context, positionModel models.TOptionPositionModel, detailModel models.TOptionSettlementDetailModel, instructionModel models.TOptionAssetInstructionModel, marginLotModel models.TOptionMarginLotModel, contract *models.TOptionContract, batchId int64, settlementNo string, deliveryPrice decimal.Decimal, now int64) (optionSettlementSummary, error) {
	cursor := int64(0)
	summary := optionSettlementSummary{}
	for {
		positions, _, err := positionModel.FindPage(ctx, models.OptionPositionPageFilter{
			ContractId: contract.Id,
			Statuses: []int64{
				int64(option.PositionStatus_POSITION_STATUS_HOLDING),
				int64(option.PositionStatus_POSITION_STATUS_EXERCISED),
				int64(option.PositionStatus_POSITION_STATUS_EXPIRED),
			},
		}, cursor, 100)
		if err != nil {
			return optionSettlementSummary{}, err
		}
		if len(positions) == 0 {
			return summary, validateOptionSettlementBalance(contract.Id, summary)
		}
		for _, position := range positions {
			cursor = position.Id
			qty := position.PositionQty
			payoff := optionSettlementPayoff(contract, deliveryPrice, qty)
			changeAmount := decimal.Zero
			exerciseFee := decimal.Zero
			if position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) {
				if contract.IsAutoExercise == int64(common.YesNo_YES_NO_YES) || position.Status == int64(option.PositionStatus_POSITION_STATUS_EXERCISED) {
					exerciseFee = payoff.Mul(contract.ExerciseFeeRate).Round(16)
					changeAmount = payoff.Sub(exerciseFee)
					summary.totalCredit = summary.totalCredit.Add(payoff)
				}
			} else if position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) {
				if contract.IsAutoExercise == int64(common.YesNo_YES_NO_YES) {
					changeAmount = payoff.Neg()
					summary.totalDebit = summary.totalDebit.Add(payoff)
				}
			}

			position.PositionQty = decimal.Zero
			position.AvailableQty = decimal.Zero
			position.FrozenQty = decimal.Zero
			position.PositionValue = decimal.Zero
			position.MarginAmount = decimal.Zero
			position.MaintenanceMargin = decimal.Zero
			position.UnrealizedPnl = decimal.Zero
			position.ExerciseableQty = decimal.Zero
			position.RealizedPnl = position.RealizedPnl.Add(changeAmount)
			position.Status = int64(option.PositionStatus_POSITION_STATUS_SETTLED)
			position.LastCalcTime = now
			position.UpdateTimes = now
			if err := positionModel.Update(ctx, position); err != nil {
				return optionSettlementSummary{}, err
			}
			direction := option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_ABANDON
			instructionNo := ""
			if changeAmount.IsPositive() {
				direction = option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_CREDIT
				instructionNo = fmt.Sprintf("%s-P%d-CREDIT", settlementNo, position.Id)
				if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
					TenantId: position.TenantId, InstructionNo: instructionNo,
					BizNo: settlementNo, PositionId: position.Id, UserId: position.UserId, AccountId: position.AccountId,
					Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
					Coin:   contract.SettleCoin, Amount: changeAmount, StepNo: 2,
					Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
					ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
					CreateTimes:          now, UpdateTimes: now,
				}); err != nil {
					return optionSettlementSummary{}, err
				}
				summary.instructionCount++
			} else if position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) {
				direction = option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_DEBIT
				if !payoff.IsPositive() {
					direction = option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_ABANDON
				}
				var count int64
				instructionNo, count, err = createShortSettlementInstructions(
					ctx, instructionModel, marginLotModel, contract, position, settlementNo, payoff, now,
				)
				if err != nil {
					return optionSettlementSummary{}, err
				}
				summary.instructionCount += count
			}
			if exerciseFee.IsPositive() {
				if contract.FeeUserId <= 0 || contract.FeeAccountId <= 0 {
					return optionSettlementSummary{}, fmt.Errorf("option exercise fee account is missing: contractId=%d", contract.Id)
				}
				feeInstructionNo := fmt.Sprintf("%s-P%d-FEE", settlementNo, position.Id)
				if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
					TenantId: position.TenantId, InstructionNo: feeInstructionNo,
					BizNo: settlementNo, PositionId: position.Id,
					UserId: contract.FeeUserId, AccountId: contract.FeeAccountId,
					Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
					Coin:   contract.SettleCoin, Amount: exerciseFee, StepNo: 2,
					Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
					ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
					CreateTimes:          now, UpdateTimes: now,
				}); err != nil {
					return optionSettlementSummary{}, err
				}
				summary.instructionCount++
				if instructionNo == "" {
					instructionNo = feeInstructionNo
				}
			}
			if _, err := detailModel.Insert(ctx, &models.TOptionSettlementDetail{
				TenantId: position.TenantId, BatchId: batchId, BatchNo: settlementNo,
				ContractId: contract.Id, PositionId: position.Id, UserId: position.UserId, AccountId: position.AccountId,
				Side: position.Side, Quantity: qty, Payoff: payoff, Direction: int64(direction),
				InstructionNo: instructionNo, CreateTimes: now,
			}); err != nil {
				return optionSettlementSummary{}, err
			}
		}
		if len(positions) < 100 {
			return summary, validateOptionSettlementBalance(contract.Id, summary)
		}
	}
}

func createShortSettlementInstructions(ctx context.Context, instructionModel models.TOptionAssetInstructionModel, marginLotModel models.TOptionMarginLotModel, contract *models.TOptionContract, position *models.TOptionPosition, settlementNo string, payoff decimal.Decimal, now int64) (string, int64, error) {
	lots, err := marginLotModel.FindActiveByPosition(ctx, position.TenantId, position.Id)
	if err != nil {
		return "", 0, err
	}
	remainingPayoff := payoff
	firstInstructionNo := ""
	count := int64(0)
	for _, lot := range lots {
		availableMargin := decimal.Max(lot.RemainingMargin.Sub(lot.PendingMargin), decimal.Zero)
		deduct := decimal.Min(availableMargin, remainingPayoff)
		if deduct.IsPositive() {
			instructionNo := fmt.Sprintf("%s-P%d-L%d-MARGIN-DEDUCT", settlementNo, position.Id, lot.Id)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: position.TenantId, InstructionNo: instructionNo,
				BizNo: settlementNo, PositionId: position.Id, MarginLotId: lot.Id,
				UserId: position.UserId, AccountId: position.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
				TargetBizNo: lot.FreezeBizNo, Coin: contract.SettleCoin, Amount: deduct, StepNo: 1,
				Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return "", 0, err
			}
			if firstInstructionNo == "" {
				firstInstructionNo = instructionNo
			}
			count++
			remainingPayoff = remainingPayoff.Sub(deduct)
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
		}
		release := availableMargin.Sub(deduct)
		if release.IsPositive() {
			instructionNo := fmt.Sprintf("%s-P%d-L%d-MARGIN-RELEASE", settlementNo, position.Id, lot.Id)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: position.TenantId, InstructionNo: instructionNo,
				BizNo: settlementNo, PositionId: position.Id, MarginLotId: lot.Id,
				UserId: position.UserId, AccountId: position.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: lot.FreezeBizNo, Coin: contract.SettleCoin, Amount: release, StepNo: 2,
				Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return "", 0, err
			}
			if firstInstructionNo == "" {
				firstInstructionNo = instructionNo
			}
			count++
			if !deduct.IsPositive() {
				lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
			}
		}
		lot.PendingMargin = lot.PendingMargin.Add(deduct).Add(release)
		lot.RemainingQuantity = decimal.Zero
		lot.UpdateTimes = now
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return "", 0, err
		}
	}
	if remainingPayoff.IsPositive() {
		instructionNo := fmt.Sprintf("%s-P%d-DEBIT-AVAILABLE", settlementNo, position.Id)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: position.TenantId, InstructionNo: instructionNo,
			BizNo: settlementNo, PositionId: position.Id, UserId: position.UserId, AccountId: position.AccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE),
			Coin:   contract.SettleCoin, Amount: remainingPayoff, StepNo: 1,
			Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return "", 0, err
		}
		if firstInstructionNo == "" {
			firstInstructionNo = instructionNo
		}
		count++
	}
	return firstInstructionNo, count, nil
}

func validateOptionSettlementBalance(contractId int64, summary optionSettlementSummary) error {
	if !summary.totalCredit.Equal(summary.totalDebit) {
		return fmt.Errorf("option settlement is not balanced: contractId=%d credit=%s debit=%s", contractId, summary.totalCredit, summary.totalDebit)
	}
	return nil
}
