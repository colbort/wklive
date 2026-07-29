package tasklogic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateCrossAccountRisk(t *testing.T) {
	equity, available, risk := calculateCrossAccountRisk(
		decimal.NewFromInt(100),
		decimal.NewFromInt(70),
		decimal.NewFromInt(30),
		decimal.NewFromInt(-20),
		decimal.NewFromInt(11),
	)
	if !equity.Equal(decimal.NewFromInt(110)) ||
		!available.Equal(decimal.NewFromInt(50)) ||
		!risk.Equal(decimal.RequireFromString("0.1")) {
		t.Fatalf("unexpected cross risk: equity=%s available=%s risk=%s", equity, available, risk)
	}
}

func TestCalculateCrossAccountRiskCapsNonPositiveEquity(t *testing.T) {
	equity, available, risk := calculateCrossAccountRisk(
		decimal.NewFromInt(10),
		decimal.NewFromInt(5),
		decimal.Zero,
		decimal.NewFromInt(-15),
		decimal.NewFromInt(1),
	)
	if !equity.Equal(decimal.NewFromInt(-5)) ||
		!available.Equal(decimal.NewFromInt(-10)) ||
		!risk.Equal(decimal.RequireFromString(crossMarginRiskRateMax)) {
		t.Fatalf("unexpected insolvent cross risk: equity=%s available=%s risk=%s", equity, available, risk)
	}
}

func TestCrossMarginProjectionSourceBindsAllVersions(t *testing.T) {
	base := crossMarginProjectionSource(1, 2, "USDT", 3, 4, 5)
	if len(base) != 51 {
		t.Fatalf("unexpected source length: %d", len(base))
	}
	tests := []string{
		crossMarginProjectionSource(9, 2, "USDT", 3, 4, 5),
		crossMarginProjectionSource(1, 9, "USDT", 3, 4, 5),
		crossMarginProjectionSource(1, 2, "BTC", 3, 4, 5),
		crossMarginProjectionSource(1, 2, "USDT", 9, 4, 5),
		crossMarginProjectionSource(1, 2, "USDT", 3, 9, 5),
		crossMarginProjectionSource(1, 2, "USDT", 3, 4, 9),
	}
	for _, changed := range tests {
		if changed == base {
			t.Fatal("projection source did not bind one of its immutable inputs")
		}
	}
}
