package risk

import (
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestEvaluatePortfolioSpreadHasBoundedScenarioLoss(t *testing.T) {
	longCall, longMarket := portfolioTestContract(1, 100, 8)
	shortCall, shortMarket := portfolioTestContract(2, 110, 4)
	result, err := EvaluatePortfolio([]PortfolioLeg{
		{Contract: longCall, Market: longMarket, LongQuantity: decimal.NewFromInt(1)},
		{Contract: shortCall, Market: shortMarket, ShortQuantity: decimal.NewFromInt(1)},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	// The debit spread's marked cost is 4 and its worst expiry loss is 4.
	if !result.ScenarioLoss.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("unexpected spread scenario loss: %s", result.ScenarioLoss)
	}
	if result.Requirement.LessThan(result.ScenarioLoss) {
		t.Fatalf("requirement cannot be below scenario loss: %+v", result)
	}
}

func TestEvaluatePortfolioDoesNotOffsetDifferentExpiries(t *testing.T) {
	longCall, longMarket := portfolioTestContract(1, 100, 8)
	shortCall, shortMarket := portfolioTestContract(2, 100, 8)
	shortCall.ExpireTime++
	result, err := EvaluatePortfolio([]PortfolioLeg{
		{Contract: longCall, Market: longMarket, LongQuantity: decimal.NewFromInt(1)},
		{Contract: shortCall, Market: shortMarket, ShortQuantity: decimal.NewFromInt(1)},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Requirement.IsPositive() {
		t.Fatal("different expiries must not receive a zero-risk offset")
	}
}

func TestEvaluatePortfolioRejectsInconsistentUnderlyingPrices(t *testing.T) {
	call1, market1 := portfolioTestContract(1, 100, 8)
	call2, market2 := portfolioTestContract(2, 110, 4)
	market2.UnderlyingPrice = decimal.NewFromInt(120)
	_, err := EvaluatePortfolio([]PortfolioLeg{
		{Contract: call1, Market: market1, LongQuantity: decimal.NewFromInt(1)},
		{Contract: call2, Market: market2, ShortQuantity: decimal.NewFromInt(1)},
	}, false)
	if err == nil {
		t.Fatal("expected inconsistent underlying price error")
	}
}

func portfolioTestContract(id, strike, mark int64) (*models.TOptionContract, *models.TOptionMarket) {
	return &models.TOptionContract{
			Id: id, UnderlyingSymbol: "BTCUSDT", SettleCoin: "USDT", ExpireTime: 1000,
			OptionType:  int64(option.OptionType_OPTION_TYPE_CALL),
			StrikePrice: decimal.NewFromInt(strike), Multiplier: decimal.NewFromInt(1),
			InitialMarginRate:     decimal.RequireFromString("0.2"),
			MaintenanceMarginRate: decimal.RequireFromString("0.1"),
			MinMarginRate:         decimal.RequireFromString("0.05"),
		}, &models.TOptionMarket{
			ContractId: id, UnderlyingPrice: decimal.NewFromInt(100),
			MarkPrice: decimal.NewFromInt(mark), SnapshotTime: 100,
		}
}
