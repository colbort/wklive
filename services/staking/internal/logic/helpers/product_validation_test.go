package helpers

import (
	"testing"

	"wklive/proto/common"
	"wklive/proto/staking"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
)

func validStakeProduct() *models.TStakeProduct {
	return &models.TStakeProduct{
		TenantId: 1, ProductNo: "USDT-30D", ProductName: "USDT 30D",
		ProductType: int64(staking.ProductType_PRODUCT_TYPE_FIXED), CoinSymbol: "USDT", RewardCoinSymbol: "USDT",
		Apr: decimal.RequireFromString("12.5"), LockDays: 30,
		MinAmount: decimal.NewFromInt(10), MaxAmount: decimal.NewFromInt(1000), StepAmount: decimal.NewFromInt(10),
		TotalAmount: decimal.NewFromInt(10000), UserLimitAmount: decimal.NewFromInt(1000),
		InterestMode: int64(staking.InterestMode_INTEREST_MODE_DAILY), RewardMode: int64(staking.RewardMode_REWARD_MODE_DAILY),
		AllowEarlyRedeem: int64(common.YesNo_YES_NO_YES), EarlyRedeemRate: decimal.RequireFromString("1.5"),
		Status: int64(staking.ProductStatus_PRODUCT_STATUS_ENABLE),
	}
}

func TestValidateStakeProduct(t *testing.T) {
	if err := ValidateStakeProduct(validStakeProduct()); err != nil {
		t.Fatalf("valid product rejected: %v", err)
	}
	tests := []struct {
		name string
		edit func(*models.TStakeProduct)
	}{
		{"quota below user limit", func(p *models.TStakeProduct) { p.TotalAmount = decimal.NewFromInt(999) }},
		{"user limit below max", func(p *models.TStakeProduct) { p.UserLimitAmount = decimal.NewFromInt(999) }},
		{"invalid step", func(p *models.TStakeProduct) { p.MinAmount = decimal.NewFromInt(15) }},
		{"fixed without lock", func(p *models.TStakeProduct) { p.LockDays = 0 }},
		{"disabled early redeem with fee", func(p *models.TStakeProduct) { p.AllowEarlyRedeem = int64(common.YesNo_YES_NO_NO) }},
		{"quota below staked", func(p *models.TStakeProduct) { p.StakedAmount = decimal.NewFromInt(10001) }},
		{"interest and reward mode mismatch", func(p *models.TStakeProduct) { p.InterestMode = int64(staking.InterestMode_INTEREST_MODE_MATURITY) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validStakeProduct()
			tt.edit(p)
			if err := ValidateStakeProduct(p); err == nil {
				t.Fatal("invalid product accepted")
			}
		})
	}
}
