package tasks

import (
	"context"
	"testing"

	"wklive/services/market/internal/config"
	"wklive/services/market/internal/market/types"
)

func TestParseExternalQuotes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		adapter   string
		symbol    string
		raw       string
		price     string
		timestamp int64
	}{
		{
			name:      "binance spot",
			adapter:   externalQuoteAdapterBinanceSpot,
			symbol:    "BTCUSDT",
			raw:       `[{"p":"64401.12000000","T":1785320000001}]`,
			price:     "64401.12000000",
			timestamp: 1785320000001,
		},
		{
			name:      "binance futures",
			adapter:   externalQuoteAdapterBinanceFutures,
			symbol:    "BTCUSDT",
			raw:       `[{"s":"BTCUSDT","p":"64402.10","T":1785320000002}]`,
			price:     "64402.10",
			timestamp: 1785320000002,
		},
		{
			name:      "okx",
			adapter:   externalQuoteAdapterOKXSpot,
			symbol:    "BTC-USDT",
			raw:       `{"code":"0","data":[{"instId":"BTC-USDT","last":"64403.2","ts":"1785320000003"}]}`,
			price:     "64403.2",
			timestamp: 1785320000003,
		},
		{
			name:      "bybit",
			adapter:   externalQuoteAdapterBybitSpot,
			symbol:    "BTCUSDT",
			raw:       `{"retCode":0,"result":{"list":[{"symbol":"BTCUSDT","price":"64404.30","time":"1785320000004"}]}}`,
			price:     "64404.30",
			timestamp: 1785320000004,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			value, err := parseExternalQuote(
				testCase.adapter,
				testCase.symbol,
				[]byte(testCase.raw),
			)
			if err != nil {
				t.Fatal(err)
			}
			if value.price != testCase.price || value.sourceTimestamp != testCase.timestamp {
				t.Fatalf("unexpected value: %+v", value)
			}
		})
	}
}

func TestParseExternalQuoteRejectsIdentityAndValueMismatch(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name    string
		adapter string
		symbol  string
		raw     string
	}{
		{
			name:    "wrong Binance symbol",
			adapter: externalQuoteAdapterBinanceSpot,
			symbol:  "BTCUSDT",
			raw:     `[{"s":"ETHUSDT","p":"1","T":1}]`,
		},
		{
			name:    "negative OKX price",
			adapter: externalQuoteAdapterOKXSpot,
			symbol:  "BTC-USDT",
			raw:     `{"code":"0","data":[{"instId":"BTC-USDT","last":"-1","ts":"1"}]}`,
		},
		{
			name:    "missing Bybit trade",
			adapter: externalQuoteAdapterBybitSpot,
			symbol:  "BTCUSDT",
			raw:     `{"retCode":0,"result":{"list":[]}}`,
		},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if _, err := parseExternalQuote(
				testCase.adapter,
				testCase.symbol,
				[]byte(testCase.raw),
			); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestExternalQuoteRunnerRequiresThreeIndependentProviders(t *testing.T) {
	t.Parallel()
	configs := []config.ExternalQuoteSourceConf{
		testExternalQuoteConfig("binance-spot", "BINANCE", externalQuoteAdapterBinanceSpot),
		testExternalQuoteConfig("binance-futures", "BINANCE", externalQuoteAdapterBinanceFutures),
		testExternalQuoteConfig("okx-spot", "OKX", externalQuoteAdapterOKXSpot),
	}
	_, err := newExternalQuoteRunner(
		configs,
		func(context.Context, types.ClientMessage, *types.QuotePayload) error { return nil },
	)
	if err == nil || err.Error() != "external quote configuration requires at least three independent providers" {
		t.Fatalf("unexpected error: %v", err)
	}
	configs = append(configs, testExternalQuoteConfig(
		"bybit-spot",
		"BYBIT",
		externalQuoteAdapterBybitSpot,
	))
	runner, err := newExternalQuoteRunner(
		configs,
		func(context.Context, types.ClientMessage, *types.QuotePayload) error { return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(runner.sources) != 4 || distinctExternalQuoteProviders(runner.sources) != 3 {
		t.Fatalf("unexpected sources: %d", len(runner.sources))
	}
}

func TestExternalQuoteRunnerAllowsAuthorityForMultipleSymbols(t *testing.T) {
	t.Parallel()
	configs := []config.ExternalQuoteSourceConf{
		testExternalQuoteConfig("binance-public", "BINANCE", externalQuoteAdapterBinanceSpot),
		testExternalQuoteConfig("okx-public", "OKX", externalQuoteAdapterOKXSpot),
		testExternalQuoteConfig("bybit-public", "BYBIT", externalQuoteAdapterBybitSpot),
	}
	secondSymbol := configs[0]
	secondSymbol.Symbol = "ETHUSDT"
	secondSymbol.UpstreamSymbol = "ETHUSDT"
	configs = append(configs, secondSymbol)

	runner, err := newExternalQuoteRunner(
		configs,
		func(context.Context, types.ClientMessage, *types.QuotePayload) error { return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(runner.sources) != 4 {
		t.Fatalf("unexpected sources: %d", len(runner.sources))
	}
}

func TestExternalQuoteRunnerRejectsDuplicateTargetIdentity(t *testing.T) {
	t.Parallel()
	configs := []config.ExternalQuoteSourceConf{
		testExternalQuoteConfig("binance-public", "BINANCE", externalQuoteAdapterBinanceSpot),
		testExternalQuoteConfig("okx-public", "OKX", externalQuoteAdapterOKXSpot),
		testExternalQuoteConfig("bybit-public", "BYBIT", externalQuoteAdapterBybitSpot),
	}
	configs = append(configs, configs[0])

	_, err := newExternalQuoteRunner(
		configs,
		func(context.Context, types.ClientMessage, *types.QuotePayload) error { return nil },
	)
	if err == nil || err.Error() != "duplicate external quote source: authority=binance-public category=crypto market=BINANCE symbol=BTCUSDT" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExternalQuoteEndpointAllowlist(t *testing.T) {
	t.Parallel()
	config := testExternalQuoteConfig("binance-spot", "BINANCE", externalQuoteAdapterBinanceSpot)
	config.BaseURL = "https://example.com/api/v3/aggTrades"
	if _, err := normalizeExternalQuoteSource(config); err == nil {
		t.Fatal("expected endpoint allowlist rejection")
	}
	config.BaseURL = "http://api.binance.com/api/v3/aggTrades"
	if _, err := normalizeExternalQuoteSource(config); err == nil {
		t.Fatal("expected non-TLS endpoint rejection")
	}
}

func testExternalQuoteConfig(
	authority string,
	provider string,
	adapter string,
) config.ExternalQuoteSourceConf {
	endpoints := map[string]string{
		externalQuoteAdapterBinanceSpot:    "https://api.binance.com/api/v3/aggTrades",
		externalQuoteAdapterBinanceFutures: "https://fapi.binance.com/fapi/v1/aggTrades",
		externalQuoteAdapterOKXSpot:        "https://www.okx.com/api/v5/market/ticker",
		externalQuoteAdapterBybitSpot:      "https://api.bybit.com/v5/market/recent-trade",
	}
	upstreamSymbol := "BTCUSDT"
	if adapter == externalQuoteAdapterOKXSpot {
		upstreamSymbol = "BTC-USDT"
	}
	return config.ExternalQuoteSourceConf{
		Enabled:         true,
		Authority:       authority,
		ProviderCode:    provider,
		Adapter:         adapter,
		BaseURL:         endpoints[adapter],
		CategoryCode:    "crypto",
		Market:          provider,
		Symbol:          "BTCUSDT",
		UpstreamSymbol:  upstreamSymbol,
		IntervalMs:      1000,
		TimeoutMs:       3000,
		MaxSourceAgeMs:  30000,
		MaxFutureSkewMs: 5000,
	}
}
