package tasklogic

import (
	"testing"

	"wklive/proto/trade"
)

func TestFillPositionAuditDifferences(t *testing.T) {
	valid := &contractFillPositionAudit{
		PositionSide: int64(trade.PositionSide_POSITION_SIDE_LONG),
		FillQty:      auditDecimal("2"),
		FillFee:      auditDecimal("0.2"),
		HistoryCount: 1,
		ProjectedQty: auditDecimal("2"),
		ProjectedFee: auditDecimal("0.2"),
	}
	if differences := fillPositionAuditDifferences(valid); len(differences) != 0 {
		t.Fatalf("valid projection reported differences: %v", differences)
	}

	reversal := *valid
	reversal.PositionSide = int64(trade.PositionSide_POSITION_SIDE_NET)
	reversal.HistoryCount = 2
	if differences := fillPositionAuditDifferences(&reversal); len(differences) != 0 {
		t.Fatalf("valid NET reversal reported differences: %v", differences)
	}

	invalid := *valid
	invalid.HistoryCount = 0
	invalid.ProjectedQty = auditDecimal("0")
	invalid.ProjectedFee = auditDecimal("0")
	if differences := fillPositionAuditDifferences(&invalid); len(differences) < 3 {
		t.Fatalf("missing history was not fully detected: %v", differences)
	}
}
