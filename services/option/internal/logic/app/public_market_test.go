package applogic

import (
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestBuildOptionChainRowsPairsLegsAndExposesOIQuality(t *testing.T) {
	contracts := []*models.TOptionContract{
		{Id: 11, OptionType: int64(option.OptionType_OPTION_TYPE_CALL), StrikePrice: decimal.RequireFromString("100")},
		{Id: 12, OptionType: int64(option.OptionType_OPTION_TYPE_PUT), StrikePrice: decimal.RequireFromString("100")},
		{Id: 21, OptionType: int64(option.OptionType_OPTION_TYPE_CALL), StrikePrice: decimal.RequireFromString("110")},
	}
	markets := []*models.TOptionMarket{
		{ContractId: 11, MarkPrice: decimal.RequireFromString("5")},
		{ContractId: 12, MarkPrice: decimal.RequireFromString("6")},
	}
	trades := []*models.OptionTradeStatistics{
		{
			ContractId: 11,
			Volume:     decimal.RequireFromString("3.5"),
			Turnover:   decimal.RequireFromString("17.5"),
			TradeCount: 2,
		},
	}
	interests := []*models.OptionOpenInterest{
		{
			ContractId: 11,
			LongQty:    decimal.RequireFromString("8"),
			ShortQty:   decimal.RequireFromString("7"),
			AsOf:       1234,
		},
		{
			ContractId: 12,
			LongQty:    decimal.RequireFromString("4"),
			ShortQty:   decimal.RequireFromString("4"),
			AsOf:       1235,
		},
	}

	rows := buildOptionChainRows(contracts, markets, trades, interests, 1000, 2000)
	if len(rows) != 2 {
		t.Fatalf("expected 2 strike rows, got %d", len(rows))
	}
	if rows[0].StrikePrice != "100" || rows[0].Call == nil || rows[0].Put == nil {
		t.Fatalf("first strike was not paired: %+v", rows[0])
	}
	if rows[1].StrikePrice != "110" || rows[1].Call == nil || rows[1].Put != nil {
		t.Fatalf("missing put leg must remain explicit: %+v", rows[1])
	}
	callStats := rows[0].Call.Statistics
	if callStats.Volume_24H != "3.5" || callStats.Turnover_24H != "17.5" ||
		callStats.TradeCount_24H != 2 {
		t.Fatalf("unexpected 24h statistics: %+v", callStats)
	}
	if callStats.OiBalanced || callStats.OpenInterest != "8" ||
		callStats.LongOpenInterest != "8" || callStats.ShortOpenInterest != "7" {
		t.Fatalf("imbalanced OI must be visible and conservatively use the larger side: %+v", callStats)
	}
	if !rows[0].Put.Statistics.OiBalanced || rows[0].Put.Statistics.OpenInterest != "4" {
		t.Fatalf("balanced OI was not represented correctly: %+v", rows[0].Put.Statistics)
	}
}

func TestToOrderBookProtoPreservesAggregation(t *testing.T) {
	levels := []*models.OptionOrderBookLevel{
		{
			Price:      decimal.RequireFromString("10.25"),
			Qty:        decimal.RequireFromString("7.5"),
			OrderCount: 3,
		},
	}
	got := toOrderBookProto(levels)
	if len(got) != 1 || got[0].Price != "10.25" || got[0].Qty != "7.5" || got[0].OrderCount != 3 {
		t.Fatalf("unexpected order-book conversion: %+v", got)
	}
}

func TestSimpleBookIsolationProbeFailsClosed(t *testing.T) {
	if simpleBookIsolationViolation([]*models.OptionOrderBookLevel{{
		Price: decimal.NewFromInt(10), Qty: decimal.NewFromInt(1),
		OrderCount: 1, ComboOrderCount: 1,
	}}) != true {
		t.Fatal("combo shadow contribution must fail the public book closed")
	}
	if simpleBookIsolationViolation([]*models.OptionOrderBookLevel{{
		Price: decimal.NewFromInt(10), Qty: decimal.NewFromInt(1),
		OrderCount: 1,
	}}) {
		t.Fatal("simple-only aggregation must remain publishable")
	}
}

func TestValidatePublicOptionChainContractsRejectsDuplicateLeg(t *testing.T) {
	contracts := []*models.TOptionContract{
		{Id: 1, OptionType: int64(option.OptionType_OPTION_TYPE_CALL), StrikePrice: decimal.NewFromInt(100)},
		{Id: 2, OptionType: int64(option.OptionType_OPTION_TYPE_CALL), StrikePrice: decimal.NewFromInt(100)},
	}
	if err := validatePublicOptionChainContracts(contracts); err == nil {
		t.Fatal("duplicate strike/type leg must fail instead of being silently overwritten")
	}
}
