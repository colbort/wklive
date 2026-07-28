package tasklogic

import (
	"testing"

	"wklive/proto/trade"

	"github.com/shopspring/decimal"
)

func auditDecimal(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}

func TestOrderFillAuditDifferences(t *testing.T) {
	valid := &contractOrderFillAudit{
		Status:            int64(trade.OrderStatus_ORDER_STATUS_FILLED),
		OrderQty:          auditDecimal("2"),
		OrderFilledQty:    auditDecimal("2"),
		OrderFilledAmount: auditDecimal("201"),
		OrderAvgPrice:     auditDecimal("100.5"),
		OrderFee:          auditDecimal("0.2"),
		FillCount:         2,
		FillQty:           auditDecimal("2"),
		FillAmount:        auditDecimal("201"),
		FillAvgPrice:      auditDecimal("100.5"),
		FillFee:           auditDecimal("0.2"),
	}
	if differences := orderFillAuditDifferences(valid); len(differences) != 0 {
		t.Fatalf("valid audit reported differences: %v", differences)
	}

	invalid := *valid
	invalid.FillQty = auditDecimal("1")
	invalid.OrderCanceledQty = auditDecimal("1")
	invalid.OrderFee = auditDecimal("0.3")
	differences := orderFillAuditDifferences(&invalid)
	if len(differences) < 3 {
		t.Fatalf("expected quantity, fee and terminal invariant differences, got %v", differences)
	}
}

func TestOrderFillAuditInverseAverage(t *testing.T) {
	row := &contractOrderFillAudit{
		OrderQty:          auditDecimal("2"),
		OrderFilledQty:    auditDecimal("2"),
		OrderFilledAmount: auditDecimal("0.03"),
		OrderAvgPrice:     auditDecimal("100"),
		OrderFee:          decimal.Zero,
		FillQty:           auditDecimal("2"),
		FillAmount:        auditDecimal("0.03"),
		FillAvgPrice:      auditDecimal("100"),
		FillFee:           decimal.Zero,
	}
	if differences := orderFillAuditDifferences(row); len(differences) != 0 {
		t.Fatalf("inverse aggregate with matching harmonic average failed: %v", differences)
	}
}

func TestOrderFillAuditAverageUsesConfiguredPriceScale(t *testing.T) {
	row := &contractOrderFillAudit{
		PriceScale:        2,
		OrderQty:          auditDecimal("60"),
		OrderFilledQty:    auditDecimal("60"),
		OrderFilledAmount: auditDecimal("6000"),
		OrderAvgPrice:     auditDecimal("55000"),
		FillQty:           auditDecimal("60"),
		FillAmount:        auditDecimal("6000"),
		FillAvgPrice:      auditDecimal("54999.999999999999999541666666666667"),
	}
	if differences := orderFillAuditDifferences(row); len(differences) != 0 {
		t.Fatalf("sub-price-precision average drift reported as mismatch: %v", differences)
	}

	row.FillAvgPrice = auditDecimal("55000.01")
	differences := orderFillAuditDifferences(row)
	if len(differences) != 1 || differences[0] != "avg_price" {
		t.Fatalf("material average mismatch was not reported: %v", differences)
	}
}
