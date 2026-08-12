package config

import (
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestMarketRuntimeConfAcceptsLegacyConfigWithoutQuoteRecoveryFields(t *testing.T) {
	const input = `
ReconcileIntervalMinutes: 5
ReconcileWindowBars: 30
GapScanIntervalMinutes: 60
RepairBatchSize: 10
BuildingBucketTtlMinutes: 120
WsKlineStaleSeconds: 30
`

	var runtime MarketRuntimeConf
	if err := conf.LoadFromYamlBytes([]byte(input), &runtime); err != nil {
		t.Fatalf("legacy runtime config must remain valid: %v", err)
	}
	if runtime.QuoteHealthCheckSeconds != 0 || runtime.QuoteRecoveryBatchSize != 0 {
		t.Fatalf("missing optional recovery fields must keep zero values: %+v", runtime)
	}
}

func TestTraderMadeConfIsBackwardCompatibleAndOptional(t *testing.T) {
	var legacy struct {
		TraderMade TraderMadeConf `json:",optional"`
	}
	if err := conf.LoadFromYamlBytes([]byte("Existing: value\n"), &legacy); err != nil {
		t.Fatalf("legacy config without TraderMade must remain valid: %v", err)
	}
	if legacy.TraderMade.APIKey != "" || legacy.TraderMade.StreamingAPIKey != "" {
		t.Fatalf("missing TraderMade config must remain disabled: %+v", legacy.TraderMade)
	}

	const input = `
ApiURL: https://marketdata.tradermade.com/api/v1
WSURL: wss://stream.tradermade.com/feedAdv
APIKey: rest-key
StreamingAPIKey: stream-key
EnableLadder: true
`
	var configured TraderMadeConf
	if err := conf.LoadFromYamlBytes([]byte(input), &configured); err != nil {
		t.Fatalf("TraderMade config must load: %v", err)
	}
	if configured.APIKey != "rest-key" || configured.StreamingAPIKey != "stream-key" || !configured.EnableLadder {
		t.Fatalf("unexpected TraderMade config: %+v", configured)
	}
}

func TestTwelveDataConfIsBackwardCompatibleAndOptional(t *testing.T) {
	type optionalConfig struct {
		TwelveData TwelveDataConf `json:",optional"`
	}
	var legacy optionalConfig
	if err := conf.LoadFromYamlBytes([]byte("{}\n"), &legacy); err != nil {
		t.Fatalf("legacy config without TwelveData must remain valid: %v", err)
	}
	if legacy.TwelveData.APIKey != "" {
		t.Fatalf("missing TwelveData config must remain disabled: %+v", legacy.TwelveData)
	}
	var configured TwelveDataConf
	if err := conf.LoadFromYamlBytes([]byte("ApiURL: https://api.twelvedata.com\nWSURL: wss://ws.twelvedata.com/v1/quotes/price\nAPIKey: key\nRestRateLimitPerMinute: 8\nRestRateLimitBurst: 8\nRestWarmMaxSymbols: 8\n"), &configured); err != nil {
		t.Fatalf("TwelveData config must load: %v", err)
	}
	if configured.APIKey != "key" || configured.RestWarmMaxSymbols != 8 {
		t.Fatalf("unexpected TwelveData config: %+v", configured)
	}
}
