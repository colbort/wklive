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
