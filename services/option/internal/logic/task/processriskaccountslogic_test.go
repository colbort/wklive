package tasklogic

import (
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestOptionTaskSellerMargin(t *testing.T) {
	contract := &models.TOptionContract{
		OptionType:            int64(option.OptionType_OPTION_TYPE_CALL),
		StrikePrice:           decimal.NewFromInt(120),
		Multiplier:            decimal.NewFromInt(1),
		MaintenanceMarginRate: decimal.RequireFromString("0.15"),
		MinMarginRate:         decimal.RequireFromString("0.08"),
	}
	got := optionTaskSellerMargin(
		contract, decimal.NewFromInt(100), decimal.NewFromInt(5), decimal.NewFromInt(2),
	)
	if !got.Equal(decimal.NewFromInt(16)) {
		t.Fatalf("maintenance margin = %s, want 16", got)
	}
}

func TestValidateLiquidationPlanBalance(t *testing.T) {
	plan := &optionLiquidationPlan{
		quantity:      decimal.NewFromInt(2),
		takeoverCost:  decimal.NewFromInt(90),
		fee:           decimal.NewFromInt(5),
		totalRequired: decimal.NewFromInt(95),
		collateral:    decimal.NewFromInt(80),
	}
	if err := validateLiquidationPlanBalance(plan, decimal.NewFromInt(15)); err != nil {
		t.Fatalf("balanced plan rejected: %v", err)
	}
	if err := validateLiquidationPlanBalance(plan, decimal.NewFromInt(14)); err == nil {
		t.Fatal("unbalanced plan accepted")
	}
}
