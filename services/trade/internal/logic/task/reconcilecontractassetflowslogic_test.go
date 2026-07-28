package tasklogic

import (
	"testing"

	"wklive/proto/asset"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestAssetFlowMatchesInstruction(t *testing.T) {
	instruction := &models.TTradeSettlementInstruction{
		TenantId:      1,
		InstructionNo: "FND-1-ASSET",
		UserId:        2,
		Action:        int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS),
		Asset:         "USDT",
		Amount:        decimal.RequireFromString("12.5"),
	}
	flow := &asset.AssetFlow{
		FlowNo:       "FLOW-1",
		TenantId:     1,
		UserId:       2,
		Coin:         "USDT",
		BizType:      asset.BizType_BIZ_TYPE_TRADE,
		BizNo:        "FND-1-ASSET",
		OpType:       asset.AssetOpType_ASSET_OP_TYPE_SUB,
		ChangeAmount: "12.500000000000000000",
	}
	if !assetFlowMatchesInstruction(instruction, flow) {
		t.Fatal("matching immutable Asset flow was rejected")
	}

	flow.ChangeAmount = "12.6"
	if assetFlowMatchesInstruction(instruction, flow) {
		t.Fatal("amount mismatch was accepted")
	}
	flow.ChangeAmount = "12.5"
	flow.OpType = asset.AssetOpType_ASSET_OP_TYPE_ADD
	if assetFlowMatchesInstruction(instruction, flow) {
		t.Fatal("operation mismatch was accepted")
	}
	flow.OpType = asset.AssetOpType_ASSET_OP_TYPE_SUB
	flow.UserId++
	if assetFlowMatchesInstruction(instruction, flow) {
		t.Fatal("identity mismatch was accepted")
	}
}

func TestFeeReconciliationAcceptsFrozenOrAvailableDebit(t *testing.T) {
	ops := expectedAssetFlowOps(int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE))
	if _, ok := ops[asset.AssetOpType_ASSET_OP_TYPE_FREEZE_DEDUCT]; !ok {
		t.Fatal("fee from frozen balance must be accepted")
	}
	if _, ok := ops[asset.AssetOpType_ASSET_OP_TYPE_SUB]; !ok {
		t.Fatal("fee recovery from available balance must be accepted")
	}
	if _, ok := ops[asset.AssetOpType_ASSET_OP_TYPE_ADD]; ok {
		t.Fatal("fee credit must not be accepted")
	}
}

func TestSettlementAssetFlowIssueKeyIsStable(t *testing.T) {
	instruction := &models.TTradeSettlementInstruction{Id: 1, InstructionNo: "DEL-1-MARGIN"}
	first := settlementAssetFlowIssueKey(instruction)
	instruction.Id = 99
	if second := settlementAssetFlowIssueKey(instruction); first != second {
		t.Fatalf("issue key changed across database identity: %s != %s", first, second)
	}
}
