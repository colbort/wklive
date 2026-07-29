package tasklogic

import (
	"testing"

	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func completedCrossAccountLiquidationAudit() crossAccountLiquidationAudit {
	return crossAccountLiquidationAudit{
		Status:              models.ContractAccountLiquidationStatusCompleted,
		PositionCount:       2,
		StartedAt:           1,
		CompletedAt:         2,
		GrossSettlement:     decimal.NewFromInt(20),
		PositionMargin:      decimal.NewFromInt(30),
		LiquidationFee:      decimal.NewFromInt(2),
		UserCredit:          decimal.NewFromInt(18),
		ItemCount:           2,
		ClosedItemCount:     2,
		ClosedPositionCount: 2,
		HistoryCount:        2,
		CompletionEvent:     1,
		ItemPositionMargin:  decimal.NewFromInt(30),
		ItemRealizedPnl:     decimal.NewFromInt(-10),
		ItemFee:             decimal.NewFromInt(2),
		NetInstructionCount: 1,
		NetInstructionDone:  1,
		FeeInstructionCount: 1,
		FeeInstructionDone:  1,
	}
}

func TestCrossAccountLiquidationAuditMatchesCompletedSaga(t *testing.T) {
	row := completedCrossAccountLiquidationAudit()
	if matched, detail := crossAccountLiquidationAuditMatches(&row); !matched {
		t.Fatalf("completed account liquidation should match: %s", detail)
	}
}

func TestCrossAccountLiquidationAuditRejectsBrokenFacts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*crossAccountLiquidationAudit)
	}{
		{"open item", func(row *crossAccountLiquidationAudit) { row.ClosedItemCount-- }},
		{"missing history", func(row *crossAccountLiquidationAudit) { row.HistoryCount-- }},
		{"bad item total", func(row *crossAccountLiquidationAudit) {
			row.ItemRealizedPnl = row.ItemRealizedPnl.Add(decimal.NewFromInt(1))
		}},
		{"unreconciled net", func(row *crossAccountLiquidationAudit) { row.NetInstructionDone = 0 }},
		{"missing fee flow", func(row *crossAccountLiquidationAudit) { row.FeeInstructionDone = 0 }},
		{"insurance exceeds deficit", func(row *crossAccountLiquidationAudit) {
			row.InsuranceFundAmount = decimal.NewFromInt(1)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row := completedCrossAccountLiquidationAudit()
			test.mutate(&row)
			if matched, _ := crossAccountLiquidationAuditMatches(&row); matched {
				t.Fatal("broken account liquidation must fail reconciliation")
			}
		})
	}
}

func TestCrossAccountLiquidationAuditAcceptsInsuranceAndADL(t *testing.T) {
	fullInsurance := completedCrossAccountLiquidationAudit()
	fullInsurance.GrossSettlement = decimal.NewFromInt(-130)
	fullInsurance.PositionMargin = decimal.NewFromInt(30)
	fullInsurance.ItemPositionMargin = decimal.NewFromInt(30)
	fullInsurance.ItemRealizedPnl = decimal.NewFromInt(-160)
	fullInsurance.LiquidationFee = decimal.Zero
	fullInsurance.ItemFee = decimal.Zero
	fullInsurance.UserCredit = decimal.Zero
	fullInsurance.UserDebit = decimal.NewFromInt(100)
	fullInsurance.DeficitAmount = decimal.NewFromInt(30)
	fullInsurance.InsuranceFundAmount = decimal.NewFromInt(30)
	fullInsurance.FeeInstructionCount = 0
	fullInsurance.FeeInstructionDone = 0
	if matched, detail := crossAccountLiquidationAuditMatches(&fullInsurance); !matched {
		t.Fatalf("full insurance account liquidation should match: %s", detail)
	}

	partial := fullInsurance
	partial.InsuranceFundAmount = decimal.NewFromInt(10)
	partial.AdlReliefAmount = decimal.NewFromInt(20)
	partial.AdlQty = decimal.NewFromInt(2)
	partial.ItemDeficit = decimal.NewFromInt(20)
	partial.ItemAdlRelief = decimal.NewFromInt(20)
	partial.ItemAdlQty = decimal.NewFromInt(2)
	partial.AdlExecutionCount = 2
	partial.AdlCompletedCount = 2
	partial.AdlExecutionQty = decimal.NewFromInt(2)
	partial.AdlExecutionRelief = decimal.NewFromInt(20)
	if matched, detail := crossAccountLiquidationAuditMatches(&partial); !matched {
		t.Fatalf("partial insurance plus ADL account liquidation should match: %s", detail)
	}
	partial.AdlUnreconciled = 1
	if matched, _ := crossAccountLiquidationAuditMatches(&partial); matched {
		t.Fatal("unreconciled ADL asset credit must fail")
	}
}

func TestCrossAccountLiquidationAuditAcceptsDebitAndNoop(t *testing.T) {
	debit := completedCrossAccountLiquidationAudit()
	debit.GrossSettlement = decimal.NewFromInt(-5)
	debit.LiquidationFee = decimal.Zero
	debit.UserCredit = decimal.Zero
	debit.UserDebit = decimal.NewFromInt(5)
	debit.ItemRealizedPnl = decimal.NewFromInt(-35)
	debit.ItemFee = decimal.Zero
	debit.FeeInstructionCount = 0
	debit.FeeInstructionDone = 0
	if matched, detail := crossAccountLiquidationAuditMatches(&debit); !matched {
		t.Fatalf("debit settlement should match: %s", detail)
	}

	noop := crossAccountLiquidationAudit{
		Status: models.ContractAccountLiquidationStatusCompleted, CompletedAt: 1,
	}
	if matched, detail := crossAccountLiquidationAuditMatches(&noop); !matched {
		t.Fatalf("no-op recovered account liquidation should match: %s", detail)
	}
}

func TestCrossAccountLiquidationAuditDefersActiveAndFlagsManual(t *testing.T) {
	active := crossAccountLiquidationAudit{Status: models.ContractAccountLiquidationStatusClosing}
	if matched, _ := crossAccountLiquidationAuditMatches(&active); !matched {
		t.Fatal("active account liquidation must be deferred")
	}
	manual := crossAccountLiquidationAudit{Status: models.ContractAccountLiquidationStatusManualReview}
	if matched, _ := crossAccountLiquidationAuditMatches(&manual); matched {
		t.Fatal("manual-review account liquidation must create an issue")
	}
}
