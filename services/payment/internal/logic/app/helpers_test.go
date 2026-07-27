package applogic

import (
	"testing"
	"wklive/common/conv"

	"github.com/shopspring/decimal"
)

func TestPaymentAmountToTextUsesNaturalUnits(t *testing.T) {
	if got := paymentAmountToText(decimal.RequireFromString("1000.25")); got != "1000.25" {
		t.Fatalf("payment natural amount = %s, want 1000.25", got)
	}
}

func TestParsePaymentAmountRejectsInvalidDatabaseDecimal(t *testing.T) {
	for _, value := range []string{"abc", "1e3", "1000000000000000000", "1.1234567890123456789"} {
		if _, err := conv.ParseDecimalField(value); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}
