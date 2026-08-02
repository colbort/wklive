package applogic

import (
	"testing"
	"wklive/common/conv"
	"wklive/services/payment/internal/logic/helpers"

	"github.com/shopspring/decimal"
)

func TestPaymentAmountToTextUsesNaturalUnits(t *testing.T) {
	if got := helpers.PaymentAmountToText(decimal.RequireFromString("1000.25")); got != "1000.25" {
		t.Fatalf("payment natural amount = %s, want 1000.25", got)
	}
}

func TestParsePaymentAmountRejectsInvalidDatabaseDecimal(t *testing.T) {
	for _, value := range []string{"abc", "1e3", "1000000000000000000", "1.1234567890123456789"} {
		if _, err := conv.ParseBoundedDecimalField(value, 18, 18); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}
