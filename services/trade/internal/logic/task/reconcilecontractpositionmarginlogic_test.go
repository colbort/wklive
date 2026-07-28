package tasklogic

import "testing"

func TestPositionMarginAuditMatchesVerifiedAssetFlows(t *testing.T) {
	valid := &contractPositionMarginAudit{
		PositionMargin: auditDecimal("80"),
		IsolatedMargin: auditDecimal("5"),
		MarginConsumed: auditDecimal("100"),
		MarginReleased: auditDecimal("15"),
	}
	if matched, detail := positionMarginAuditMatches(valid); !matched {
		t.Fatalf("valid custody ledger rejected: %s", detail)
	}

	mismatch := *valid
	mismatch.PositionMargin = auditDecimal("79")
	if matched, _ := positionMarginAuditMatches(&mismatch); matched {
		t.Fatal("custody mismatch was accepted")
	}

	pending := mismatch
	pending.UnfinishedCount = 1
	if matched, _ := positionMarginAuditMatches(&pending); !matched {
		t.Fatal("in-flight margin instruction must defer reconciliation")
	}

	liquidation := mismatch
	liquidation.LiquidationCount = 1
	if matched, _ := positionMarginAuditMatches(&liquidation); !matched {
		t.Fatal("liquidation position must be handled by the dedicated reconciliation")
	}
}
