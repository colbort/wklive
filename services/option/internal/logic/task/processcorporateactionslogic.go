package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const corporateActionPositionBatchSize int64 = 100

type ProcessCorporateActionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessCorporateActionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessCorporateActionsLogic {
	return &ProcessCorporateActionsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// 按批次执行已复核且到达生效时间的公司行动迁移
func (l *ProcessCorporateActionsLogic) ProcessCorporateActions(
	in *option.OptionTaskReq,
) (*option.OptionTaskResp, error) {
	now := time.Now().Unix()
	actions, err := l.svcCtx.OptionCorporateActionModel.FindDue(l.ctx, in.TenantId, now, 20)
	if err != nil {
		return nil, err
	}
	for _, action := range actions {
		mappings, findErr := l.svcCtx.OptionCorporateActionContractModel.FindByAction(
			l.ctx, action.TenantId, action.Id,
		)
		if findErr != nil {
			return nil, findErr
		}
		for _, mapping := range mappings {
			if mapping.Status == int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_COMPLETED) ||
				mapping.Status == int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_MANUAL_REVIEW) {
				continue
			}
			if processErr := l.processMapping(action.Id, mapping.Id, now); processErr != nil {
				l.Errorf("corporate action mapping failed, actionId=%d mappingId=%d err=%v",
					action.Id, mapping.Id, processErr)
				if markErr := l.markMappingFailure(action.Id, mapping.Id, processErr, now); markErr != nil {
					return nil, markErr
				}
			}
		}
		if finalizeErr := l.finalizeAction(action.Id, now); finalizeErr != nil {
			return nil, finalizeErr
		}
	}
	return &option.OptionTaskResp{Base: helper.OkResp()}, nil
}

func (l *ProcessCorporateActionsLogic) processMapping(actionID, mappingID, now int64) error {
	action, err := l.svcCtx.OptionCorporateActionModel.FindOne(l.ctx, actionID)
	if err != nil {
		return err
	}
	mapping, err := l.svcCtx.OptionCorporateActionContractModel.FindOne(l.ctx, mappingID)
	if err != nil {
		return err
	}
	if action.Status != int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_APPROVED) &&
		action.Status != int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_EXECUTING) &&
		action.Status != int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_FAILED) {
		return nil
	}
	if mapping.ExecutionMode != int64(option.CorporateActionExecutionMode_CORPORATE_ACTION_EXECUTION_MODE_AUTO_CASH_SUCCESSOR) {
		return i18n.StatusError(l.ctx, i18n.OperationNotAllowed)
	}
	if mapping.Status == int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_FAILED) {
		if err := l.resumeFailedMapping(action.Id, mapping.Id, now); err != nil {
			return err
		}
		action, err = l.svcCtx.OptionCorporateActionModel.FindOne(l.ctx, actionID)
		if err != nil {
			return err
		}
		mapping, err = l.svcCtx.OptionCorporateActionContractModel.FindOne(l.ctx, mappingID)
		if err != nil {
			return err
		}
	}
	if mapping.LastPositionId == 0 {
		if err := l.validateExecutionGate(action, mapping); err != nil {
			return err
		}
		if err := l.initializeMapping(action.Id, mapping.Id, now); err != nil {
			return err
		}
		mapping, err = l.svcCtx.OptionCorporateActionContractModel.FindOne(l.ctx, mappingID)
		if err != nil {
			return err
		}
	}
	positions, err := l.svcCtx.OptionPositionModel.FindHoldingBatch(
		l.ctx, mapping.TenantId, mapping.SourceContractId, mapping.LastPositionId, corporateActionPositionBatchSize,
	)
	if err != nil {
		return err
	}
	for _, position := range positions {
		if err := l.migratePosition(action.Id, mapping.Id, position.Id, now); err != nil {
			return err
		}
	}
	if len(positions) == 0 {
		return l.completeMapping(action.Id, mapping.Id, now)
	}
	return nil
}

func (l *ProcessCorporateActionsLogic) resumeFailedMapping(actionID, mappingID, now int64) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		actionModel := models.NewTOptionCorporateActionModel(conn, l.svcCtx.Config.CacheRedis)
		mappingModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		action, err := actionModel.FindOneForUpdate(ctx, actionID)
		if err != nil {
			return err
		}
		mapping, err := mappingModel.FindOneForUpdate(ctx, mappingID)
		if err != nil {
			return err
		}
		if mapping.Status != int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_FAILED) {
			return nil
		}
		mapping.Status = int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_EXECUTING)
		mapping.LastErrorMsg = ""
		mapping.UpdateTimes = now
		if err = mappingModel.Update(ctx, mapping); err != nil {
			return err
		}
		action.Status = int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_EXECUTING)
		action.LastErrorMsg = ""
		action.UpdateTimes = now
		return actionModel.Update(ctx, action)
	})
}

func (l *ProcessCorporateActionsLogic) initializeMapping(actionID, mappingID, now int64) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		actionModel := models.NewTOptionCorporateActionModel(conn, l.svcCtx.Config.CacheRedis)
		mappingModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		action, err := actionModel.FindOneForUpdate(ctx, actionID)
		if err != nil {
			return err
		}
		mapping, err := mappingModel.FindOneForUpdate(ctx, mappingID)
		if err != nil {
			return err
		}
		if mapping.LastPositionId != 0 {
			return nil
		}
		total, err := positionModel.CountHoldingByContract(ctx, mapping.TenantId, mapping.SourceContractId)
		if err != nil {
			return err
		}
		mapping.PositionTotal = total
		mapping.Status = int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_EXECUTING)
		mapping.LastErrorMsg = ""
		mapping.UpdateTimes = now
		if err := mappingModel.Update(ctx, mapping); err != nil {
			return err
		}
		action.Status = int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_EXECUTING)
		action.LastErrorMsg = ""
		action.UpdateTimes = now
		return actionModel.Update(ctx, action)
	})
}

func (l *ProcessCorporateActionsLogic) validateExecutionGate(
	action *models.TOptionCorporateAction,
	mapping *models.TOptionCorporateActionContract,
) error {
	if action == nil || mapping == nil || action.TenantId != mapping.TenantId {
		return i18n.StatusError(l.ctx, i18n.OperationNotAllowed)
	}
	source, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, mapping.SourceContractId)
	if err != nil {
		return err
	}
	successor, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, mapping.SuccessorContractId)
	if err != nil {
		return err
	}
	halt, err := l.svcCtx.OptionTradingHaltModel.FindOne(l.ctx, mapping.HaltId)
	if err != nil {
		return err
	}
	if source.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) ||
		successor.Status != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) ||
		halt.Status != int64(option.TradingHaltStatus_TRADING_HALT_STATUS_ACTIVE) ||
		halt.Source != int64(option.TradingHaltSource_TRADING_HALT_SOURCE_CORPORATE_ACTION) {
		return i18n.StatusError(l.ctx, i18n.OperationNotAllowed)
	}
	blocked, err := models.CountCorporateActionExecutionBlockers(
		l.ctx, l.svcCtx.DB, mapping.TenantId, mapping.SourceContractId,
	)
	if err != nil {
		return err
	}
	if blocked > 0 {
		return fmt.Errorf("corporate action execution blocked by %d unfinished records", blocked)
	}
	return nil
}

func (l *ProcessCorporateActionsLogic) migratePosition(
	actionID, mappingID, positionID, now int64,
) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		actionModel := models.NewTOptionCorporateActionModel(conn, l.svcCtx.Config.CacheRedis)
		mappingModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		auditModel := models.NewTOptionCorporateActionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		marginAuditModel := models.NewTOptionCorporateActionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		marginModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionExerciseInstructionModel(conn, l.svcCtx.Config.CacheRedis)

		action, err := actionModel.FindOneForUpdate(ctx, actionID)
		if err != nil {
			return err
		}
		mapping, err := mappingModel.FindOneForUpdate(ctx, mappingID)
		if err != nil {
			return err
		}
		position, err := positionModel.FindOneForUpdate(ctx, positionID)
		if err != nil {
			return err
		}
		if existing, findErr := auditModel.FindOneByTenantIdActionContractIdSourcePositionId(
			ctx, mapping.TenantId, mapping.Id, position.Id,
		); findErr == nil {
			if existing.Status != int64(option.CorporateActionPositionStatus_CORPORATE_ACTION_POSITION_STATUS_COMPLETED) {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			if position.Id > mapping.LastPositionId {
				mapping.LastPositionId = position.Id
				mapping.UpdateTimes = now
				return mappingModel.Update(ctx, mapping)
			}
			return nil
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		if action.Status != int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_EXECUTING) ||
			mapping.Status != int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_EXECUTING) ||
			position.ContractId != mapping.SourceContractId ||
			position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) ||
			position.FrozenQty.IsPositive() {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		source, err := contractModel.FindOneForUpdate(ctx, mapping.SourceContractId)
		if err != nil {
			return err
		}
		successor, err := contractModel.FindOneForUpdate(ctx, mapping.SuccessorContractId)
		if err != nil {
			return err
		}
		conversion, err := helpers.ConvertCorporateActionPosition(
			position, source, successor, mapping.QuantityNumerator, mapping.QuantityDenominator,
		)
		if err != nil {
			return err
		}
		exerciseable, err := helpers.ExactCorporateActionQuantity(
			position.ExerciseableQty, mapping.QuantityNumerator, mapping.QuantityDenominator,
		)
		if err != nil {
			return err
		}
		successorPosition := &models.TOptionPosition{
			TenantId: position.TenantId, UserId: position.UserId, AccountId: position.AccountId,
			ContractId: successor.Id, UnderlyingSymbol: successor.UnderlyingSymbol, Side: position.Side,
			PositionQty:  conversion.SuccessorQuantity,
			AvailableQty: conversion.SuccessorAvailableQuantity, FrozenQty: decimal.Zero,
			OpenAvgPrice: conversion.SuccessorOpenAvgPrice, MarkPrice: decimal.Zero,
			PositionValue: decimal.Zero, MarginAmount: position.MarginAmount,
			MaintenanceMargin: position.MaintenanceMargin, UnrealizedPnl: decimal.Zero,
			RealizedPnl: decimal.Zero, TradeRealizedPnl: decimal.Zero,
			SettlementRealizedPnl: decimal.Zero, FeePaid: decimal.Zero, TotalReturn: decimal.Zero,
			ExerciseableQty: exerciseable,
			Status:          int64(option.PositionStatus_POSITION_STATUS_HOLDING),
			LastCalcTime:    now, CreateTimes: now, UpdateTimes: now,
		}
		result, err := positionModel.Insert(ctx, successorPosition)
		if err != nil {
			return err
		}
		successorPosition.Id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		audit := &models.TOptionCorporateActionPosition{
			TenantId: mapping.TenantId, ActionId: action.Id, ActionContractId: mapping.Id,
			SourcePositionId: position.Id, SuccessorPositionId: successorPosition.Id,
			UserId: position.UserId, AccountId: position.AccountId, Side: position.Side,
			SourceQuantity: position.PositionQty, SuccessorQuantity: conversion.SuccessorQuantity,
			SourceAvailableQuantity:      position.AvailableQty,
			SuccessorAvailableQuantity:   conversion.SuccessorAvailableQuantity,
			SourceOpenAvgPrice:           position.OpenAvgPrice,
			SuccessorOpenAvgPrice:        conversion.SuccessorOpenAvgPrice,
			SourceEffectiveMultiplier:    conversion.SourceEffectiveMultiplier,
			SuccessorEffectiveMultiplier: conversion.TargetEffectiveMultiplier,
			CostBasisBefore:              conversion.CostBasisBefore, CostBasisAfter: conversion.CostBasisAfter,
			CashDifference: decimal.Zero,
			Status:         int64(option.CorporateActionPositionStatus_CORPORATE_ACTION_POSITION_STATUS_COMPLETED),
			CompletedAt:    now, CreateTimes: now, UpdateTimes: now,
		}
		result, err = auditModel.Insert(ctx, audit)
		if err != nil {
			return err
		}
		audit.Id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		marginLots, err := marginModel.FindRemainingByPositionForUpdate(ctx, mapping.TenantId, position.Id)
		if err != nil {
			return err
		}
		for _, lot := range marginLots {
			if lot.PendingMargin.IsPositive() ||
				(lot.RemainingQuantity.IsPositive() &&
					lot.Status != int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE)) {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			successorLotQty, convertErr := helpers.ExactCorporateActionQuantity(
				lot.Quantity, mapping.QuantityNumerator, mapping.QuantityDenominator,
			)
			if convertErr != nil {
				return convertErr
			}
			successorRemaining, convertErr := helpers.ExactCorporateActionQuantity(
				lot.RemainingQuantity, mapping.QuantityNumerator, mapping.QuantityDenominator,
			)
			if convertErr != nil {
				return convertErr
			}
			_, err = marginAuditModel.Insert(ctx, &models.TOptionCorporateActionMarginLot{
				TenantId: mapping.TenantId, ActionPositionId: audit.Id, MarginLotId: lot.Id,
				SourceContractId: source.Id, SuccessorContractId: successor.Id,
				SourcePositionId: position.Id, SuccessorPositionId: successorPosition.Id,
				SourceQuantity: lot.Quantity, SuccessorQuantity: successorLotQty,
				SourceRemainingQuantity:    lot.RemainingQuantity,
				SuccessorRemainingQuantity: successorRemaining, CreateTimes: now,
			})
			if err != nil {
				return err
			}
			lot.ContractId = successor.Id
			lot.PositionId = successorPosition.Id
			lot.CorporateActionPositionId = audit.Id
			lot.Quantity = successorLotQty
			lot.RemainingQuantity = successorRemaining
			lot.UpdateTimes = now
			if err = marginModel.Update(ctx, lot); err != nil {
				return err
			}
		}
		instruction, findErr := instructionModel.FindLatestByPositionForUpdate(
			ctx, position.TenantId, position.Id,
		)
		if findErr == nil && instruction.Status == int64(option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE) {
			instruction.Status = int64(option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_SUPERSEDED)
			instruction.UpdateTimes = now
			if err = instructionModel.Update(ctx, instruction); err != nil {
				return err
			}
		} else if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		position.AvailableQty = decimal.Zero
		position.FrozenQty = decimal.Zero
		position.MarginAmount = decimal.Zero
		position.MaintenanceMargin = decimal.Zero
		position.UnrealizedPnl = decimal.Zero
		position.ExerciseableQty = decimal.Zero
		position.Status = int64(option.PositionStatus_POSITION_STATUS_MIGRATED)
		position.UpdateTimes = now
		if err = positionModel.Update(ctx, position); err != nil {
			return err
		}
		mapping.PositionCompleted++
		mapping.LastPositionId = position.Id
		mapping.LastErrorMsg = ""
		mapping.UpdateTimes = now
		return mappingModel.Update(ctx, mapping)
	})
}

func (l *ProcessCorporateActionsLogic) completeMapping(actionID, mappingID, now int64) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		actionModel := models.NewTOptionCorporateActionModel(conn, l.svcCtx.Config.CacheRedis)
		mappingModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		action, err := actionModel.FindOneForUpdate(ctx, actionID)
		if err != nil {
			return err
		}
		mapping, err := mappingModel.FindOneForUpdate(ctx, mappingID)
		if err != nil {
			return err
		}
		if mapping.Status == int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_COMPLETED) {
			return nil
		}
		if mapping.PositionCompleted != mapping.PositionTotal || mapping.PositionFailed != 0 {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		source, err := contractModel.FindOneForUpdate(ctx, mapping.SourceContractId)
		if err != nil {
			return err
		}
		halt, err := haltModel.FindOneForUpdate(ctx, mapping.HaltId)
		if err != nil {
			return err
		}
		source.Status = int64(option.ContractStatus_CONTRACT_STATUS_OFFLINE)
		source.UpdateTimes = now
		if err = contractModel.Update(ctx, source); err != nil {
			return err
		}
		halt.Status = int64(option.TradingHaltStatus_TRADING_HALT_STATUS_LIFTED)
		halt.ActiveKey = "HALT:" + halt.HaltNo
		halt.LiftedAt = now
		halt.LiftedBy = 0
		halt.LiftReason = "corporate action completed; source contract offline"
		halt.LastErrorMsg = ""
		halt.UpdateTimes = now
		if err = haltModel.Update(ctx, halt); err != nil {
			return err
		}
		mapping.Status = int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_COMPLETED)
		mapping.LastErrorMsg = ""
		mapping.UpdateTimes = now
		if err = mappingModel.Update(ctx, mapping); err != nil {
			return err
		}
		_, err = eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: source.TenantId, ContractId: source.Id,
			EventType: "CORPORATE_ACTION_MIGRATED", Reason: action.EventNo,
			Detail: fmt.Sprintf("actionId=%d mappingId=%d successorContractId=%d positions=%d",
				action.Id, mapping.Id, mapping.SuccessorContractId, mapping.PositionCompleted),
			OperatorId: 0, CreateTimes: now,
		})
		return err
	})
}

func (l *ProcessCorporateActionsLogic) markMappingFailure(
	actionID, mappingID int64, cause error, now int64,
) error {
	message := cause.Error()
	if len(message) > 500 {
		message = message[:500]
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		actionModel := models.NewTOptionCorporateActionModel(conn, l.svcCtx.Config.CacheRedis)
		mappingModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		action, err := actionModel.FindOneForUpdate(ctx, actionID)
		if err != nil {
			return err
		}
		mapping, err := mappingModel.FindOneForUpdate(ctx, mappingID)
		if err != nil {
			return err
		}
		if mapping.Status == int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_COMPLETED) {
			return nil
		}
		mapping.RetryCount++
		mapping.LastErrorMsg = message
		mapping.UpdateTimes = now
		action.LastErrorMsg = message
		action.UpdateTimes = now
		if mapping.RetryCount >= 3 {
			mapping.Status = int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_MANUAL_REVIEW)
			action.Status = int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_MANUAL_REVIEW)
		} else {
			mapping.Status = int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_FAILED)
			action.Status = int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_FAILED)
		}
		if err = mappingModel.Update(ctx, mapping); err != nil {
			return err
		}
		return actionModel.Update(ctx, action)
	})
}

func (l *ProcessCorporateActionsLogic) finalizeAction(actionID, now int64) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		actionModel := models.NewTOptionCorporateActionModel(conn, l.svcCtx.Config.CacheRedis)
		mappingModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		action, err := actionModel.FindOneForUpdate(ctx, actionID)
		if err != nil {
			return err
		}
		if action.Status == int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_MANUAL_REVIEW) ||
			action.Status == int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_COMPLETED) {
			return nil
		}
		mappings, err := mappingModel.FindByAction(ctx, action.TenantId, action.Id)
		if err != nil {
			return err
		}
		for _, mapping := range mappings {
			if mapping.Status != int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_COMPLETED) {
				return nil
			}
		}
		action.Status = int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_COMPLETED)
		action.CompletedAt = now
		action.LastErrorMsg = ""
		action.UpdateTimes = now
		return actionModel.Update(ctx, action)
	})
}
