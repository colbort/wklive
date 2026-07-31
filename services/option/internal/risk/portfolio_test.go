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
	}, false, portfolioTestConfig())
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
	}, false, portfolioTestConfig())
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
	}, false, portfolioTestConfig())
	if err == nil {
		t.Fatal("expected inconsistent underlying price error")
	}
}

func TestEvaluatePortfolioAddsConcentrationAndLiquidityAtExactBoundary(t *testing.T) {
	shortCall, market := portfolioTestContract(1, 100, 8)
	config := portfolioTestConfig()
	config.ConcentrationThreshold = decimal.NewFromInt(100)
	config.ConcentrationAddonRate = decimal.RequireFromString("0.1")
	config.LiquidityAddonRate = decimal.RequireFromString("0.02")
	result, err := EvaluatePortfolio([]PortfolioLeg{{
		Contract: shortCall, Market: market, ShortQuantity: decimal.NewFromInt(1),
	}}, false, config)
	if err != nil {
		t.Fatal(err)
	}
	if !result.ConcentrationAddon.IsZero() {
		t.Fatalf("exact threshold must not add concentration margin: %s", result.ConcentrationAddon)
	}
	if !result.LiquidityAddon.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("unexpected liquidity addon: %s", result.LiquidityAddon)
	}

	config.ConcentrationThreshold = decimal.NewFromInt(50)
	increased, err := EvaluatePortfolio([]PortfolioLeg{{
		Contract: shortCall, Market: market, ShortQuantity: decimal.NewFromInt(1),
	}}, false, config)
	if err != nil {
		t.Fatal(err)
	}
	if !increased.ConcentrationAddon.Equal(decimal.NewFromInt(5)) ||
		!increased.Requirement.GreaterThan(result.Requirement) {
		t.Fatalf("concentration addon must increase requirement: before=%+v after=%+v", result, increased)
	}
}

func TestParseScenarioShocksRequiresZeroAndFiveTimesCoverage(t *testing.T) {
	if _, _, err := ParseScenarioShocks("-0.5,1"); err == nil {
		t.Fatal("scenario set without total-loss and five-times coverage must be rejected")
	}
	shocks, canonical, err := ParseScenarioShocks("4, -1, 0.2,4")
	if err != nil {
		t.Fatal(err)
	}
	if len(shocks) != 3 || canonical != "-1,0.2,4" {
		t.Fatalf("unexpected canonical shock set: %q %+v", canonical, shocks)
	}
}

func portfolioTestConfig() PortfolioConfig {
	return PortfolioConfig{
		InitialShockRate:       decimal.RequireFromString("0.2"),
		MaintenanceShockRate:   decimal.RequireFromString("0.1"),
		ScenarioShocks:         []decimal.Decimal{decimal.NewFromInt(-1), decimal.NewFromInt(4)},
		ConcentrationThreshold: decimal.NewFromInt(1000),
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
