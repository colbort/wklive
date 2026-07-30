package tasklogic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestValidateOptionSettlementBalance(t *testing.T) {
	if err := validateOptionSettlementBalance(1, optionSettlementSummary{
		totalCredit: decimal.NewFromInt(10),
		totalDebit:  decimal.NewFromInt(10),
	}); err != nil {
		t.Fatalf("balanced settlement rejected: %v", err)
	}

	if err := validateOptionSettlementBalance(1, optionSettlementSummary{
		totalCredit: decimal.NewFromInt(10),
		totalDebit:  decimal.NewFromInt(9),
	}); err == nil {
		t.Fatal("unbalanced settlement must be rejected")
	}
}
