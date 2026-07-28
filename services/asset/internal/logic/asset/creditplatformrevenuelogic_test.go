package assetlogic

import (
	"testing"

	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
)

func TestPlatformRevenueFlowMatches(t *testing.T) {
	flow := &models.TAssetPlatformFlow{
		AccountType: feeRevenueAccountType,
		OpType:      1,
		Amount:      decimal.NewFromInt(5),
		BizType:     "trade",
		SceneType:   "trade_fee",
		BizId:       42,
	}
	if !platformRevenueFlowMatches(flow, decimal.NewFromInt(5), "trade", "trade_fee", 42) {
		t.Fatal("identical platform revenue replay must match")
	}
	if platformRevenueFlowMatches(flow, decimal.NewFromInt(6), "trade", "trade_fee", 42) {
		t.Fatal("changed amount must be rejected")
	}
	if platformRevenueFlowMatches(flow, decimal.NewFromInt(5), "trade", "trade_fee", 43) {
		t.Fatal("changed business id must be rejected")
	}
}
