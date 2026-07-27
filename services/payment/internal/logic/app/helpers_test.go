package applogic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestPaymentAmountToTextUsesNaturalUnits(t *testing.T) {
	if got := paymentAmountToText(decimal.RequireFromString("1000.25")); got != "1000.25" {
		t.Fatalf("payment natural amount = %s, want 1000.25", got)
	}
}
