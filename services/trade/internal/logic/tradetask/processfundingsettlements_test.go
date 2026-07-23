package tradetasklogic

import (
	"testing"

	"github.com/shopspring/decimal"
	"wklive/services/trade/models"
)

func TestFundingDifferenceInstructionBalancesUserNet(t *testing.T) {
	action, step := fundingDifferenceInstruction(decimal.RequireFromString("0.01"))
	if action != 8 || step != 1 {
		t.Fatalf("positive user net must debit platform first: action=%d step=%d", action, step)
	}
	action, step = fundingDifferenceInstruction(decimal.RequireFromString("-0.01"))
	if action != 3 || step != 2 {
		t.Fatalf("negative user net must credit platform after payers: action=%d step=%d", action, step)
	}
}

func TestSettlementInstructionIdentityIncludesSagaBinding(t *testing.T) {
	base := &models.TTradeSettlementInstruction{TenantId: 1, InstructionNo: "i", BizType: "funding", BizId: "s", BatchNo: "b", PositionId: 2, UserId: 3, Action: 8, Asset: "USDT", Amount: decimal.NewFromInt(1), StepNo: 1}
	copy := *base
	if !sameSettlementInstructionIdentity(base, &copy) {
		t.Fatal("identical instruction rejected")
	}
	copy.PositionId++
	if sameSettlementInstructionIdentity(base, &copy) {
		t.Fatal("different position binding accepted")
	}
	copy = *base
	copy.StepNo++
	if sameSettlementInstructionIdentity(base, &copy) {
		t.Fatal("different saga step accepted")
	}
}
