package tasklogic

import (
	"testing"

	"wklive/proto/trade"

	"github.com/shopspring/decimal"
)

func completedLiquidationAudit() contractLiquidationAudit {
	return contractLiquidationAudit{
		Status:              int64(trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED),
		TriggerQty:          decimal.NewFromInt(10),
		LiquidatedQty:       decimal.NewFromInt(10),
		CompletedAt:         1,
		PositionStatus:      int64(trade.PositionStatus_POSITION_STATUS_CLOSED),
		LiquidationHistory:  1,
		CompletionEvent:     1,
		AdlExecutionCount:   2,
		AdlCompletedCount:   2,
		AdlQty:              decimal.NewFromInt(3),
		AdlExecutionQty:     decimal.NewFromInt(3),
		AdlReliefAmount:     decimal.NewFromInt(25),
		InsuranceFundAmount: decimal.NewFromInt(5),
	}
}

func TestLiquidationAuditMatchesCompletedSaga(t *testing.T) {
	row := completedLiquidationAudit()
	if matched, detail := liquidationAuditMatches(&row); !matched {
		t.Fatalf("completed liquidation should match: %s", detail)
	}
}

func TestLiquidationAuditRejectsBrokenADLAndPosition(t *testing.T) {
	row := completedLiquidationAudit()
	row.AdlUnreconciledAssets = 1
	if matched, _ := liquidationAuditMatches(&row); matched {
		t.Fatal("unreconciled ADL asset credit must fail")
	}
	row = completedLiquidationAudit()
	row.PositionMargin = decimal.NewFromInt(1)
	if matched, _ := liquidationAuditMatches(&row); matched {
		t.Fatal("uncleared bankrupt position margin must fail")
	}
}

func TestLiquidationAuditDefersActiveAndFlagsManual(t *testing.T) {
	active := contractLiquidationAudit{Status: int64(trade.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING)}
	if matched, _ := liquidationAuditMatches(&active); !matched {
		t.Fatal("active saga must be deferred")
	}
	manual := contractLiquidationAudit{Status: int64(trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW)}
	if matched, _ := liquidationAuditMatches(&manual); matched {
		t.Fatal("manual-review liquidation must be recorded as an issue")
	}
}
