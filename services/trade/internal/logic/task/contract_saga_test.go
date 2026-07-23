package tasklogic

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestDeliveryAssetStepsDebitBeforeCredit(t *testing.T) {
	steps := deliveryAssetSteps(decimal.NewFromInt(10), decimal.NewFromInt(-2), decimal.NewFromInt(1))
	if len(steps) != 3 {
		t.Fatalf("unexpected step count: %d", len(steps))
	}
	for _, step := range steps {
		if step.action == trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_PNL_LOSS && step.stepNo != 1 {
			t.Fatalf("debit must be step 1: %+v", step)
		}
		if step.action == trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE && step.stepNo != 2 {
			t.Fatalf("credit must be step 2: %+v", step)
		}
	}
}

func TestDeliveryAssetStepsOmitZeroAmounts(t *testing.T) {
	steps := deliveryAssetSteps(decimal.NewFromInt(10), decimal.Zero, decimal.Zero)
	if len(steps) != 1 || steps[0].suffix != "MARGIN" {
		t.Fatalf("unexpected zero filtering: %+v", steps)
	}
}

func TestDeliveryInstructionMustMatchImmutableStep(t *testing.T) {
	steps := deliveryAssetSteps(decimal.NewFromInt(10), decimal.NewFromInt(2), decimal.NewFromInt(1))
	item := &models.TTradeSettlementInstruction{InstructionNo: "S-MARGIN", BizId: "S", Action: 3, Amount: decimal.NewFromInt(10), StepNo: 2}
	if !matchesDeliveryAssetStep(item, steps) {
		t.Fatal("valid delivery instruction rejected")
	}
	item.Amount = decimal.NewFromInt(11)
	if matchesDeliveryAssetStep(item, steps) {
		t.Fatal("modified delivery amount accepted")
	}
}

func TestSettlementInstructionLeaseFencing(t *testing.T) {
	claimed := &models.TTradeSettlementInstruction{Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PROCESSING), UpdateTimes: 100}
	current := &models.TTradeSettlementInstruction{Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PROCESSING), UpdateTimes: 100}
	if !settlementInstructionLeaseOwned(current, claimed) {
		t.Fatal("current lease should be accepted")
	}
	current.UpdateTimes = 200
	if settlementInstructionLeaseOwned(current, claimed) {
		t.Fatal("expired worker must not own a newer lease")
	}
	current.UpdateTimes = 100
	current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)
	if settlementInstructionLeaseOwned(current, claimed) {
		t.Fatal("completed instruction must not remain lease-owned")
	}
}
