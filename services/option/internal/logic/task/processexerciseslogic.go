package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessExercisesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessExercisesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessExercisesLogic {
	return &ProcessExercisesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 处理美式主动行权的空头指派和资金清算
func (l *ProcessExercisesLogic) ProcessExercises(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_exercises", func() (*option.OptionTaskResp, error) {
		items, err := l.svcCtx.OptionExerciseModel.FindPending(l.ctx, in.TenantId, 100)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			assignments, err := l.svcCtx.OptionExerciseAssignmentModel.FindByExercise(
				l.ctx, item.TenantId, item.Id,
			)
			if err != nil {
				return nil, err
			}
			if len(assignments) > 0 {
				if err := completeExerciseIfReady(l.ctx, l.svcCtx, item.TenantId, item.ExerciseNo); err != nil {
					l.Errorf("reconcile pending option exercise failed, exerciseNo=%s err=%v", item.ExerciseNo, err)
				}
				continue
			}
			if err := l.createExerciseClearing(item); err != nil {
				l.Errorf("create option exercise clearing failed, exerciseNo=%s err=%v", item.ExerciseNo, err)
			}
		}
		return helpers.OkTaskResp(), nil
	})
}

func (l *ProcessExercisesLogic) createExerciseClearing(exercise *models.TOptionExercise) error {
	var canceledOrders []*models.TOptionOrder
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		exerciseModel := models.NewTOptionExerciseModel(conn, l.svcCtx.Config.CacheRedis)
		assignmentModel := models.NewTOptionExerciseAssignmentModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := exerciseModel.FindOneForUpdate(ctx, exercise.Id)
		if err != nil {
			return err
		}
		if current.Status != int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING) {
			return nil
		}
		existing, err := assignmentModel.FindByExercise(ctx, current.TenantId, current.Id)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return nil
		}
		contract, err := contractModel.FindOne(ctx, current.ContractId)
		if err != nil {
			return err
		}
		if contract.ExerciseStyle != int64(option.ExerciseStyle_EXERCISE_STYLE_AMERICAN) ||
			contract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) {
			return errors.New("only cash-settled American options support early exercise")
		}
		longPosition, err := positionModel.FindOneForUpdate(ctx, current.PositionId)
		if err != nil {
			return err
		}
		if longPosition.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) ||
			longPosition.FrozenQty.LessThan(current.ExerciseQty) ||
			longPosition.ExerciseableQty.LessThan(current.ExerciseQty) {
			return errors.New("early exercise long position reservation changed")
		}
		now := time.Now().Unix()
		remainingQty := current.ExerciseQty
		totalDebit := decimal.Zero
		const assignmentPageSize int64 = 500
		cursorCreateTimes := int64(0)
		cursorID := int64(0)
		for remainingQty.IsPositive() {
			shorts, err := positionModel.FindAssignableShortsPage(
				ctx, current.TenantId, current.ContractId,
				cursorCreateTimes, cursorID, assignmentPageSize,
			)
			if err != nil {
				return err
			}
			if len(shorts) == 0 {
				break
			}
			for _, candidate := range shorts {
				cursorCreateTimes = candidate.CreateTimes
				cursorID = candidate.Id
				if !remainingQty.IsPositive() {
					break
				}
				closeOrders, err := orderModel.FindActiveCloseOrdersForUpdate(
					ctx,
					candidate.TenantId, candidate.UserId, candidate.AccountId, candidate.ContractId,
				)
				if err != nil {
					return err
				}
				for _, closeOrder := range closeOrders {
					changed, err := cancelOptionSystemOrderInTx(
						ctx, orderModel, positionModel, instructionModel,
						closeOrder, "AMERICAN_EXERCISE_ASSIGNMENT", now,
					)
					if err != nil {
						return err
					}
					if changed {
						canceledOrders = append(canceledOrders, closeOrder)
					}
				}
				// The page query bounds memory and establishes deterministic FIFO candidates.
				// Re-read each candidate with a current locking read after close-order cancellation,
				// so concurrent exercises cannot assign stale available quantity.
				short, err := positionModel.FindOneForUpdate(ctx, candidate.Id)
				if err != nil {
					return err
				}
				assignQty := decimal.Min(remainingQty, short.AvailableQty)
				if !assignQty.IsPositive() {
					continue
				}
				payoff := optionSettlementPayoff(contract, current.SettlementPrice, assignQty)
				if !payoff.IsPositive() {
					return errors.New("early exercise payoff must be positive")
				}
				firstInstruction, err := createExerciseShortInstructions(
					ctx, instructionModel, marginLotModel, contract, current, short,
					assignQty, payoff, now,
				)
				if err != nil {
					return err
				}
				if _, err := assignmentModel.Insert(ctx, &models.TOptionExerciseAssignment{
					TenantId: current.TenantId, ExerciseId: current.Id, ExerciseNo: current.ExerciseNo,
					LongPositionId: longPosition.Id, ShortPositionId: short.Id,
					ShortUserId: short.UserId, ShortAccountId: short.AccountId,
					Quantity: assignQty, Payoff: payoff,
					Status:        int64(option.ExerciseAssignmentStatus_EXERCISE_ASSIGNMENT_STATUS_PENDING),
					InstructionNo: firstInstruction, CreateTimes: now, UpdateTimes: now,
				}); err != nil {
					return err
				}
				allocatedMaintenance := decimal.Zero
				if short.PositionQty.IsPositive() && short.MaintenanceMargin.IsPositive() {
					allocatedMaintenance = short.MaintenanceMargin.Mul(assignQty).
						Div(short.PositionQty).Round(16)
				}
				short.PositionQty = decimal.Max(short.PositionQty.Sub(assignQty), decimal.Zero)
				short.AvailableQty = decimal.Max(short.AvailableQty.Sub(assignQty), decimal.Zero)
				short.MaintenanceMargin = decimal.Max(
					short.MaintenanceMargin.Sub(allocatedMaintenance), decimal.Zero,
				)
				short.PositionValue = short.MarkPrice.Mul(short.PositionQty).Mul(optionMultiplier(contract)).Round(16)
				applyPositionSettlementReturn(
					short, payoff, decimal.Zero, assignQty, optionMultiplier(contract),
				)
				short.UnrealizedPnl = short.OpenAvgPrice.Sub(short.MarkPrice).
					Mul(short.PositionQty).Mul(optionMultiplier(contract)).Round(16)
				short.UpdateTimes = now
				if short.PositionQty.IsZero() {
					short.AvailableQty = decimal.Zero
					short.FrozenQty = decimal.Zero
					short.PositionValue = decimal.Zero
					short.MaintenanceMargin = decimal.Zero
					short.UnrealizedPnl = decimal.Zero
					short.Status = int64(option.PositionStatus_POSITION_STATUS_EXERCISED)
				}
				if err := positionModel.Update(ctx, short); err != nil {
					return err
				}
				totalDebit = totalDebit.Add(payoff)
				remainingQty = remainingQty.Sub(assignQty)
			}
			if int64(len(shorts)) < assignmentPageSize {
				break
			}
		}
		netCredit := current.ProfitAmount.Sub(current.Fee)
		if remainingQty.IsPositive() {
			return fmt.Errorf(
				"early exercise assignment quantity is incomplete: remaining=%s",
				remainingQty,
			)
		}
		if err := validateEarlyExerciseBalance(totalDebit, current.ProfitAmount, current.Fee, netCredit); err != nil {
			return err
		}
		if err := insertExerciseInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
			TenantId: current.TenantId, InstructionNo: current.ExerciseNo + "-LONG-CREDIT",
			BizNo: current.ExerciseNo, PositionId: longPosition.Id,
			UserId: current.UserId, AccountId: current.AccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
			Coin:   contract.SettleCoin, Amount: netCredit, StepNo: 2,
			CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		if current.Fee.IsPositive() {
			if contract.FeeUserId <= 0 || contract.FeeAccountId <= 0 {
				return errors.New("early exercise fee account is not configured")
			}
			if err := insertExerciseInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
				TenantId: current.TenantId, InstructionNo: current.ExerciseNo + "-FEE-CREDIT",
				BizNo: current.ExerciseNo, PositionId: longPosition.Id,
				UserId: contract.FeeUserId, AccountId: contract.FeeAccountId,
				Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
				Coin:   contract.SettleCoin, Amount: current.Fee, StepNo: 2,
				CreateTimes: now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		longPosition.PositionQty = decimal.Max(longPosition.PositionQty.Sub(current.ExerciseQty), decimal.Zero)
		longPosition.FrozenQty = decimal.Max(longPosition.FrozenQty.Sub(current.ExerciseQty), decimal.Zero)
		longPosition.ExerciseableQty = decimal.Max(longPosition.ExerciseableQty.Sub(current.ExerciseQty), decimal.Zero)
		longPosition.PositionValue = longPosition.MarkPrice.Mul(longPosition.PositionQty).
			Mul(optionMultiplier(contract)).Round(16)
		applyPositionSettlementReturn(
			longPosition, current.ProfitAmount, current.Fee,
			current.ExerciseQty, optionMultiplier(contract),
		)
		longPosition.UnrealizedPnl = longPosition.MarkPrice.Sub(longPosition.OpenAvgPrice).
			Mul(longPosition.PositionQty).Mul(optionMultiplier(contract)).Round(16)
		longPosition.UpdateTimes = now
		if longPosition.PositionQty.IsZero() {
			longPosition.AvailableQty = decimal.Zero
			longPosition.FrozenQty = decimal.Zero
			longPosition.ExerciseableQty = decimal.Zero
			longPosition.PositionValue = decimal.Zero
			longPosition.UnrealizedPnl = decimal.Zero
			longPosition.Status = int64(option.PositionStatus_POSITION_STATUS_EXERCISED)
		}
		return positionModel.Update(ctx, longPosition)
	})
	if err != nil {
		return err
	}
	for _, order := range canceledOrders {
		publishOptionOrderChanged(l.ctx, l.svcCtx, order)
	}
	return nil
}

func validateEarlyExerciseBalance(totalDebit, grossPayoff, fee, netCredit decimal.Decimal) error {
	if !grossPayoff.IsPositive() || fee.IsNegative() || !netCredit.IsPositive() ||
		!totalDebit.Equal(grossPayoff) || !netCredit.Add(fee).Equal(totalDebit) {
		return errors.New("early exercise clearing is not balanced")
	}
	return nil
}

func createExerciseShortInstructions(
	ctx context.Context,
	instructionModel models.TOptionAssetInstructionModel,
	marginLotModel models.TOptionMarginLotModel,
	contract *models.TOptionContract,
	exercise *models.TOptionExercise,
	short *models.TOptionPosition,
	quantity, payoff decimal.Decimal,
	now int64,
) (string, error) {
	if contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) {
		return createPortfolioExerciseShortInstructions(
			ctx, instructionModel, marginLotModel, contract, exercise, short, payoff, now,
		)
	}
	lots, err := marginLotModel.FindClosableByPosition(ctx, exercise.TenantId, short.Id)
	if err != nil {
		return "", err
	}
	quantityLeft := quantity
	payoffLeft := payoff
	firstInstruction := ""
	for _, lot := range lots {
		if !quantityLeft.IsPositive() {
			break
		}
		if lot.PendingMargin.IsPositive() {
			return "", errors.New("early exercise margin lot has pending amount")
		}
		assignedQty := decimal.Min(quantityLeft, lot.RemainingQuantity)
		if !assignedQty.IsPositive() {
			continue
		}
		allocatedMargin := lot.RemainingMargin
		if assignedQty.LessThan(lot.RemainingQuantity) {
			allocatedMargin = lot.RemainingMargin.Mul(assignedQty).Div(lot.RemainingQuantity).Round(16)
		}
		deduct := decimal.Min(allocatedMargin, payoffLeft)
		release := allocatedMargin.Sub(deduct)
		if deduct.IsPositive() {
			instructionNo := fmt.Sprintf("%s-S%d-L%d-DEDUCT", exercise.ExerciseNo, short.Id, lot.Id)
			if err := insertExerciseInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
				TenantId: exercise.TenantId, InstructionNo: instructionNo,
				BizNo: exercise.ExerciseNo, PositionId: short.Id, MarginLotId: lot.Id,
				UserId: short.UserId, AccountId: short.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
				TargetBizNo: lot.FreezeBizNo, Coin: contract.SettleCoin, Amount: deduct,
				StepNo: 1, CreateTimes: now, UpdateTimes: now,
			}); err != nil {
				return "", err
			}
			if firstInstruction == "" {
				firstInstruction = instructionNo
			}
			payoffLeft = payoffLeft.Sub(deduct)
		}
		if release.IsPositive() {
			instructionNo := fmt.Sprintf("%s-S%d-L%d-RELEASE", exercise.ExerciseNo, short.Id, lot.Id)
			if err := insertExerciseInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
				TenantId: exercise.TenantId, InstructionNo: instructionNo,
				BizNo: exercise.ExerciseNo, PositionId: short.Id, MarginLotId: lot.Id,
				UserId: short.UserId, AccountId: short.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: lot.FreezeBizNo, Coin: contract.SettleCoin, Amount: release,
				StepNo: 1, CreateTimes: now, UpdateTimes: now,
			}); err != nil {
				return "", err
			}
			if firstInstruction == "" {
				firstInstruction = instructionNo
			}
		}
		lot.RemainingQuantity = decimal.Max(lot.RemainingQuantity.Sub(assignedQty), decimal.Zero)
		lot.PendingMargin = lot.PendingMargin.Add(allocatedMargin)
		if deduct.IsPositive() {
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
		} else {
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
		}
		lot.UpdateTimes = now
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return "", err
		}
		quantityLeft = quantityLeft.Sub(assignedQty)
	}
	if quantityLeft.IsPositive() {
		return "", fmt.Errorf("insufficient margin lot quantity for exercise assignment: %s", quantityLeft)
	}
	if payoffLeft.IsPositive() {
		instructionNo := fmt.Sprintf("%s-S%d-DEBIT-AVAILABLE", exercise.ExerciseNo, short.Id)
		if err := insertExerciseInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
			TenantId: exercise.TenantId, InstructionNo: instructionNo,
			BizNo: exercise.ExerciseNo, PositionId: short.Id,
			UserId: short.UserId, AccountId: short.AccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE),
			Coin:   contract.SettleCoin, Amount: payoffLeft,
			StepNo: 1, CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return "", err
		}
		if firstInstruction == "" {
			firstInstruction = instructionNo
		}
	}
	return firstInstruction, nil
}

func createPortfolioExerciseShortInstructions(
	ctx context.Context,
	instructionModel models.TOptionAssetInstructionModel,
	marginLotModel models.TOptionMarginLotModel,
	contract *models.TOptionContract,
	exercise *models.TOptionExercise,
	short *models.TOptionPosition,
	payoff decimal.Decimal,
	now int64,
) (string, error) {
	lots, err := marginLotModel.FindPortfolioActiveByAccount(
		ctx, exercise.TenantId, short.UserId, short.AccountId, contract.SettleCoin,
	)
	if err != nil {
		return "", err
	}
	payoffLeft := payoff
	firstInstruction := ""
	for _, lot := range lots {
		if !payoffLeft.IsPositive() {
			break
		}
		if lot.PendingMargin.IsPositive() {
			return "", errors.New("portfolio exercise margin lot has pending amount")
		}
		deduct := decimal.Min(lot.RemainingMargin, payoffLeft)
		if !deduct.IsPositive() {
			continue
		}
		instructionNo := fmt.Sprintf("%s-S%d-PL%d-DEDUCT", exercise.ExerciseNo, short.Id, lot.Id)
		if err := insertExerciseInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
			TenantId: exercise.TenantId, InstructionNo: instructionNo,
			BizNo: exercise.ExerciseNo, PositionId: short.Id, MarginLotId: lot.Id,
			UserId: short.UserId, AccountId: short.AccountId,
			Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
			TargetBizNo: lot.FreezeBizNo, Coin: contract.SettleCoin, Amount: deduct,
			StepNo: 1, CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return "", err
		}
		if firstInstruction == "" {
			firstInstruction = instructionNo
		}
		lot.PendingMargin = lot.PendingMargin.Add(deduct)
		lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
		lot.UpdateTimes = now
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return "", err
		}
		payoffLeft = payoffLeft.Sub(deduct)
	}
	if payoffLeft.IsPositive() {
		instructionNo := fmt.Sprintf("%s-S%d-DEBIT-AVAILABLE", exercise.ExerciseNo, short.Id)
		if err := insertExerciseInstruction(ctx, instructionModel, &models.TOptionAssetInstruction{
			TenantId: exercise.TenantId, InstructionNo: instructionNo,
			BizNo: exercise.ExerciseNo, PositionId: short.Id,
			UserId: short.UserId, AccountId: short.AccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE),
			Coin:   contract.SettleCoin, Amount: payoffLeft,
			StepNo: 1, CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return "", err
		}
		if firstInstruction == "" {
			firstInstruction = instructionNo
		}
	}
	return firstInstruction, nil
}

func insertExerciseInstruction(
	ctx context.Context,
	model models.TOptionAssetInstructionModel,
	item *models.TOptionAssetInstruction,
) error {
	item.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING)
	item.ReconciliationStatus = int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING)
	_, err := model.Insert(ctx, item)
	return err
}
