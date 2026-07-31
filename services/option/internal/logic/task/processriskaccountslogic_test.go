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

func TestOptionRiskEquityUsesSignedMarkValue(t *testing.T) {
	tests := []struct {
		name           string
		assetTotal     string
		netOptionValue string
		want           string
	}{
		{
			name:           "long premium already paid",
			assetTotal:     "95",
			netOptionValue: "5",
			want:           "100",
		},
		{
			name:           "short premium already received",
			assetTotal:     "105",
			netOptionValue: "-5",
			want:           "100",
		},
		{
			name:           "hedged long and short",
			assetTotal:     "100",
			netOptionValue: "0",
			want:           "100",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optionRiskEquity(
				decimal.RequireFromString(tt.assetTotal),
				decimal.RequireFromString(tt.netOptionValue),
			)
			if !got.Equal(decimal.RequireFromString(tt.want)) {
				t.Fatalf("equity=%s want=%s", got, tt.want)
			}
		})
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

func TestLiquidationDeficitResolution(t *testing.T) {
	deficit := decimal.NewFromInt(10)
	tests := []struct {
		name      string
		insurance decimal.Decimal
		backstop  decimal.Decimal
		want      option.LiquidationDeficitResolution
	}{
		{
			name: "insurance", insurance: decimal.NewFromInt(10),
			want: option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_INSURANCE_FUND,
		},
		{
			name: "backstop", backstop: decimal.NewFromInt(10),
			want: option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_PLATFORM_BACKSTOP,
		},
		{
			name: "combined", insurance: decimal.NewFromInt(4), backstop: decimal.NewFromInt(6),
			want: option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_INSURANCE_AND_BACKSTOP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := liquidationDeficitResolution(deficit, tt.insurance, tt.backstop)
			if got != tt.want {
				t.Fatalf("resolution=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestLiquidationCoverageFlowNumbersAreReplaySafe(t *testing.T) {
	liq := &models.TOptionLiquidation{LiquidationNo: "OLQ-1", InsuranceAttempt: 2}
	if got := liquidationInsuranceFlowNo(liq, false); got != "OLQ-1-INSURANCE-A2" {
		t.Fatalf("manual insurance flow=%q", got)
	}
	firstInsurance := liquidationInsuranceFlowNo(liq, true)
	firstBackstop := liquidationBackstopFlowNo(liq)
	liq.InsuranceAttempt++
	if got := liquidationInsuranceFlowNo(liq, true); got != firstInsurance {
		t.Fatalf("backstop-mode insurance key changed across retry: %q -> %q", firstInsurance, got)
	}
	if got := liquidationBackstopFlowNo(liq); got != firstBackstop {
		t.Fatalf("backstop key changed across retry: %q -> %q", firstBackstop, got)
	}
}
