package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

const accountLiquidationBizType = "cross_liquidation"

type RetryAccountLiquidationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryAccountLiquidationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryAccountLiquidationLogic {
	return &RetryAccountLiquidationLogic{
		ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx),
	}
}

func (l *RetryAccountLiquidationLogic) RetryAccountLiquidation(
	in *trade.RetryAccountLiquidationReq,
) (*trade.CommonResp, error) {
	tenantID := helpers.AdminTenantID(l.ctx, in.GetTenantId())
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil || operatorID <= 0 {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "missing admin operator identity")}, nil
	}
	reason := strings.TrimSpace(in.GetReason())
	if reason == "" || len(reason) > 500 {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "retry reason is required and must not exceed 500 bytes")}, nil
	}
	if !l.svcCtx.Config.AutomaticLiquidation.Enabled {
		return &trade.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "automatic liquidation gate is disabled")}, nil
	}
	eventNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "TRE", "")
	if err != nil {
		return nil, err
	}
	notFound, invalidStatus, invalidFacts := false, false, false
	now := utils.NowMillis()
	err = l.svcCtx.TransactionModel.TransactOnce(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		parentModel := tx.ContractAccountLiquidation
		parent, findErr := parentModel.FindOneForUpdate(ctx, in.GetId())
		if errors.Is(findErr, models.ErrNotFound) || (findErr == nil && parent.TenantId != tenantID) {
			notFound = true
			return nil
		}
		if findErr != nil {
			return findErr
		}
		if parent.Status != models.ContractAccountLiquidationStatusManualReview {
			invalidStatus = true
			return nil
		}
		items, findErr := tx.ContractAccountLiquidationItem.FindByLiquidation(ctx, tenantID, parent.Id, true)
		if findErr != nil {
			return findErr
		}
		nextStatus := models.ContractAccountLiquidationStatusPending
		if len(items) > 0 {
			nextStatus, invalidFacts, findErr = resetAccountLiquidationInstructions(
				ctx, tx, parent, items, now,
			)
			if findErr != nil || invalidFacts {
				return findErr
			}
		}
		parent.Status = nextStatus
		parent.Reason = "manual retry requested: " + reason
		parent.Version++
		parent.UpdateTimes = now
		if err = parentModel.Update(ctx, parent); err != nil {
			return err
		}
		_, err = tx.BizTradeEvent.Insert(ctx, &models.TBizTradeEvent{
			TenantId: tenantID, EventNo: eventNo,
			EventType: "CROSS_ACCOUNT_LIQUIDATION_RETRY_REQUESTED",
			BizId:     parent.LiquidationNo, BizType: accountLiquidationBizType,
			UserId: parent.UserId, OperatorId: operatorID,
			Source:        int64(trade.SourceType_SOURCE_TYPE_ADMIN),
			EventStatus:   int64(trade.EventStatus_EVENT_STATUS_PENDING),
			MaxRetryCount: 20, NextRetryAt: now,
			Payload:     helpers.NormalizeTradeEventJSON(reason),
			CreateTimes: now, UpdateTimes: now,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	switch {
	case notFound:
		return &trade.CommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, "account liquidation not found")}, nil
	case invalidStatus:
		return &trade.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "only manual-review account liquidations can be retried")}, nil
	case invalidFacts:
		return &trade.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, "account liquidation settlement facts are incomplete")}, nil
	default:
		return &trade.CommonResp{Base: helper.OkResp()}, nil
	}
}

func resetAccountLiquidationInstructions(
	ctx context.Context, tx *models.TransactionModels,
	parent *models.TContractAccountLiquidation,
	items []*models.TContractAccountLiquidationItem,
	now int64,
) (int64, bool, error) {
	instructionModel := tx.TradeSettlementInstruction
	netRequired := parent.UserCredit.IsPositive() || parent.UserDebit.IsPositive()
	netDone := !netRequired
	if netRequired {
		net, err := instructionModel.FindOneByTenantIdInstructionNo(ctx, parent.TenantId, parent.LiquidationNo+"-NET")
		if errors.Is(err, models.ErrNotFound) {
			return 0, true, nil
		}
		if err != nil {
			return 0, false, err
		}
		if net.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			netDone = true
		} else if isRetryableSettlementInstruction(net.Status) {
			resetSettlementInstruction(net, now)
			if err = instructionModel.Update(ctx, net); err != nil {
				return 0, false, err
			}
		}
	}
	if parent.LiquidationFee.IsPositive() {
		fee, err := instructionModel.FindOneByTenantIdInstructionNo(ctx, parent.TenantId, parent.LiquidationNo+"-FEE")
		if errors.Is(err, models.ErrNotFound) {
			return 0, true, nil
		}
		if err != nil {
			return 0, false, err
		}
		if isRetryableSettlementInstruction(fee.Status) {
			resetSettlementInstruction(fee, now)
			if err = instructionModel.Update(ctx, fee); err != nil {
				return 0, false, err
			}
		}
	}
	if !netDone {
		return models.ContractAccountLiquidationStatusAssetSettling, false, nil
	}
	hasADLFacts := false
	adlModel := tx.ContractAdlExecution
	for _, item := range items {
		if item.DeficitAmount.IsPositive() || item.BankruptcyPrice.IsPositive() ||
			item.AdlReliefAmount.IsPositive() || item.AdlQty.IsPositive() {
			hasADLFacts = true
		}
		executions, err := adlModel.FindByLiquidation(ctx, parent.TenantId, -item.Id)
		if err != nil {
			return 0, false, err
		}
		if len(executions) > 0 {
			hasADLFacts = true
		}
		for _, execution := range executions {
			if execution.Status != 4 && execution.Status != 5 {
				continue
			}
			nextExecutionStatus := int64(1)
			if execution.AssetCredit.IsPositive() {
				instruction, findErr := instructionModel.FindOneByTenantIdInstructionNo(
					ctx, parent.TenantId, execution.ExecutionNo,
				)
				if errors.Is(findErr, models.ErrNotFound) {
					return 0, true, nil
				}
				if findErr != nil {
					return 0, false, findErr
				}
				if instruction.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
					nextExecutionStatus = 2
				} else if isRetryableSettlementInstruction(instruction.Status) {
					resetSettlementInstruction(instruction, now)
					if findErr = instructionModel.Update(ctx, instruction); findErr != nil {
						return 0, false, findErr
					}
				} else {
					return 0, true, nil
				}
			}
			execution.Status = nextExecutionStatus
			execution.LastErrorMsg = ""
			execution.UpdateTimes = now
			if err = adlModel.Update(ctx, execution); err != nil {
				return 0, false, err
			}
		}
	}
	return accountLiquidationRetryStage(parent, hasADLFacts), false, nil
}

func accountLiquidationRetryStage(parent *models.TContractAccountLiquidation, hasADLFacts bool) int64 {
	if parent == nil || !parent.DeficitAmount.IsPositive() {
		return models.ContractAccountLiquidationStatusClosing
	}
	covered := parent.InsuranceFundAmount.Add(parent.AdlReliefAmount)
	if covered.GreaterThanOrEqual(parent.DeficitAmount) ||
		parent.DeficitAmount.Sub(covered).Abs().LessThanOrEqual(decimal.New(1, -12)) {
		return models.ContractAccountLiquidationStatusClosing
	}
	if hasADLFacts || parent.InsuranceFundAmount.IsPositive() {
		return models.ContractAccountLiquidationStatusADL
	}
	return models.ContractAccountLiquidationStatusInsuranceFund
}

func isRetryableSettlementInstruction(status int64) bool {
	return status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED) ||
		status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW)
}

func resetSettlementInstruction(instruction *models.TTradeSettlementInstruction, now int64) {
	instruction.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING)
	instruction.RetryCount = 0
	instruction.NextRetryAt = now
	instruction.LastErrorMsg = ""
	instruction.UpdateTimes = now
}
