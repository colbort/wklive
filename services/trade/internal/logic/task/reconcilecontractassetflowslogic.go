package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const settlementAssetFlowCheck = "SETTLEMENT_ASSET_FLOW"

type ReconcileContractAssetFlowsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReconcileContractAssetFlowsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReconcileContractAssetFlowsLogic {
	return &ReconcileContractAssetFlowsLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ReconcileContractAssetFlowsLogic) Process(tenantID int64) error {
	var result error
	cursor := int64(0)
	for processed := 0; processed < 1000; {
		rows, err := l.svcCtx.TradeSettlementInstrModel.FindSuccessUnreconciled(l.ctx, tenantID, cursor, 100)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		progressed := false
		for _, instruction := range rows {
			cursor = instruction.Id
			processed++
			matched, flowNo, actual, reconcileErr := l.reconcileInstruction(instruction)
			if reconcileErr != nil {
				result = errors.Join(result, reconcileErr)
			}
			if matched {
				instruction.AssetFlowNo = flowNo
				instruction.ReconciledAt = utils.NowMillis()
				instruction.UpdateTimes = instruction.ReconciledAt
				if err = l.svcCtx.TradeSettlementInstrModel.Update(l.ctx, instruction); err != nil {
					return err
				}
				if err = l.svcCtx.ContractReconcileIssueModel.ResolveByKey(
					l.ctx, instruction.TenantId, settlementAssetFlowIssueKey(instruction),
					"matching Asset flow observed", instruction.ReconciledAt,
				); err != nil {
					return err
				}
				progressed = true
				continue
			}
			now := utils.NowMillis()
			expected := expectedAssetFlowSummary(instruction)
			detail := "Asset flow is missing or does not match immutable settlement instruction"
			if reconcileErr != nil {
				detail = reconcileErr.Error()
			}
			if err = l.recordContractReconciliationFinding(&models.TContractReconciliationIssue{
				TenantId:      instruction.TenantId,
				IssueKey:      settlementAssetFlowIssueKey(instruction),
				CheckType:     settlementAssetFlowCheck,
				BizType:       instruction.BizType,
				BizNo:         instruction.InstructionNo,
				InstructionId: instruction.Id,
				ExpectedValue: expected,
				ActualValue:   actual,
				Detail:        detail,
				FirstSeenAt:   now,
				LastSeenAt:    now,
				CreateTimes:   now,
				UpdateTimes:   now,
			}); err != nil {
				return err
			}
		}
		if !progressed && len(rows) < 100 {
			break
		}
	}
	if err := l.reconcileOrderFills(tenantID); err != nil {
		result = errors.Join(result, fmt.Errorf("Order/Fill reconciliation: %w", err))
	}
	if err := l.reconcileReservations(tenantID); err != nil {
		result = errors.Join(result, fmt.Errorf("Reservation/Asset Freeze reconciliation: %w", err))
	}
	if err := l.reconcileFillPositionHistories(tenantID); err != nil {
		result = errors.Join(result, fmt.Errorf("Fill/Position History reconciliation: %w", err))
	}
	if err := l.reconcilePositionMarginCustody(tenantID); err != nil {
		result = errors.Join(result, fmt.Errorf("Position margin/Asset custody reconciliation: %w", err))
	}
	if err := l.reconcileLiquidations(tenantID); err != nil {
		result = errors.Join(result, fmt.Errorf("Liquidation/insurance/ADL reconciliation: %w", err))
	}
	if err := l.reconcileCrossAccountLiquidations(tenantID); err != nil {
		result = errors.Join(result, fmt.Errorf("Cross account liquidation reconciliation: %w", err))
	}
	return result
}

func (l *ReconcileContractAssetFlowsLogic) reconcileInstruction(instruction *models.TTradeSettlementInstruction) (bool, string, string, error) {
	if instruction == nil {
		return false, "", "nil instruction", errors.New("cannot reconcile nil settlement instruction")
	}
	resp, err := l.queryInstructionAssetFlows(instruction, instruction.InstructionNo)
	if err != nil {
		return false, "", "asset query failed", err
	}
	if len(resp.GetData()) == 0 {
		if refundBizNo, ok := legacySecondsRefundBizNo(instruction); ok {
			refundResp, refundErr := l.queryInstructionAssetFlows(instruction, refundBizNo)
			if refundErr != nil {
				return false, "", "legacy seconds refund query failed", refundErr
			}
			if len(refundResp.GetData()) == 1 {
				flow := refundResp.GetData()[0]
				actual := actualAssetFlowSummary(flow)
				if assetFlowMatchesInstructionBizNo(instruction, flow, refundBizNo) {
					return true, flow.GetFlowNo(), actual, nil
				}
				return false, flow.GetFlowNo(), actual, nil
			}
		}
	}
	if len(resp.GetData()) != 1 {
		return false, "", fmt.Sprintf("flow_count=%d", len(resp.GetData())), nil
	}
	flow := resp.GetData()[0]
	actual := actualAssetFlowSummary(flow)
	if !assetFlowMatchesInstruction(instruction, flow) {
		return false, flow.GetFlowNo(), actual, nil
	}
	return true, flow.GetFlowNo(), actual, nil
}

func (l *ReconcileContractAssetFlowsLogic) queryInstructionAssetFlows(instruction *models.TTradeSettlementInstruction, bizNo string) (*asset.PageAssetFlowsResp, error) {
	resp, err := l.svcCtx.AssetAdminClient.PageAssetFlows(l.ctx, &asset.PageAssetFlowsReq{
		TenantId: instruction.TenantId,
		UserId:   instruction.UserId,
		Coin:     instruction.Asset,
		BizType:  asset.BizType_BIZ_TYPE_TRADE,
		BizNo:    bizNo,
		Page:     &common.PageReq{Limit: 2},
	})
	if err != nil {
		return nil, fmt.Errorf("query Asset flow for %s: %w", bizNo, err)
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 {
		return nil, fmt.Errorf("Asset flow query rejected for %s", bizNo)
	}
	return resp, nil
}

func settlementAssetFlowIssueKey(instruction *models.TTradeSettlementInstruction) string {
	return "ASSET_FLOW:" + instruction.InstructionNo
}

func expectedAssetFlowOps(action int64) map[asset.AssetOpType]struct{} {
	ops := make(map[asset.AssetOpType]struct{})
	switch trade.SettlementInstructionAction(action) {
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN,
		trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_ADJUST_MARGIN:
		ops[asset.AssetOpType_ASSET_OP_TYPE_FREEZE_DEDUCT] = struct{}{}
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN:
		ops[asset.AssetOpType_ASSET_OP_TYPE_UNFREEZE] = struct{}{}
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE,
		trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_POST_PNL,
		trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_MARGIN:
		ops[asset.AssetOpType_ASSET_OP_TYPE_ADD] = struct{}{}
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE:
		ops[asset.AssetOpType_ASSET_OP_TYPE_FREEZE_DEDUCT] = struct{}{}
		ops[asset.AssetOpType_ASSET_OP_TYPE_SUB] = struct{}{}
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS:
		ops[asset.AssetOpType_ASSET_OP_TYPE_SUB] = struct{}{}
	}
	return ops
}

func assetFlowMatchesInstruction(instruction *models.TTradeSettlementInstruction, flow *asset.AssetFlow) bool {
	return assetFlowMatchesInstructionBizNo(instruction, flow, instruction.InstructionNo)
}

func assetFlowMatchesInstructionBizNo(instruction *models.TTradeSettlementInstruction, flow *asset.AssetFlow, bizNo string) bool {
	if instruction == nil || flow == nil ||
		flow.GetTenantId() != instruction.TenantId ||
		flow.GetUserId() != instruction.UserId ||
		flow.GetCoin() != instruction.Asset ||
		flow.GetBizNo() != bizNo ||
		flow.GetBizType() != asset.BizType_BIZ_TYPE_TRADE {
		return false
	}
	ops := expectedAssetFlowOps(instruction.Action)
	if _, ok := ops[flow.GetOpType()]; !ok {
		return false
	}
	amount, err := decimal.NewFromString(flow.GetChangeAmount())
	return err == nil && amount.Abs().Equal(instruction.Amount)
}

func legacySecondsRefundBizNo(instruction *models.TTradeSettlementInstruction) (string, bool) {
	if instruction == nil ||
		trade.SettlementInstructionAction(instruction.Action) != trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN ||
		instruction.ReservationNo == "" ||
		instruction.InstructionNo != instruction.ReservationNo+"-RELEASE" {
		return "", false
	}
	return instruction.ReservationNo + "-SECONDS-REFUND", true
}

func expectedAssetFlowSummary(instruction *models.TTradeSettlementInstruction) string {
	names := make([]string, 0, 2)
	for op := range expectedAssetFlowOps(instruction.Action) {
		names = append(names, op.String())
	}
	sort.Strings(names)
	return fmt.Sprintf("user=%d asset=%s amount=%s ops=%s",
		instruction.UserId, instruction.Asset, instruction.Amount, strings.Join(names, "|"))
}

func actualAssetFlowSummary(flow *asset.AssetFlow) string {
	if flow == nil {
		return "missing"
	}
	return fmt.Sprintf("flow=%s user=%d asset=%s amount=%s op=%s",
		flow.GetFlowNo(), flow.GetUserId(), flow.GetCoin(), flow.GetChangeAmount(), flow.GetOpType())
}
