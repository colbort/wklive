package validation

import (
	"testing"

	"wklive/proto/common"
)

func TestAuthoritativeQuoteSources(t *testing.T) {
	for _, value := range []string{
		"crypto:BA:BTCUSDT",
		"crypto:BA:BTCUSDT,crypto:BB:BTCUSDT",
	} {
		if err := AuthoritativeQuoteSources("source", value); err != nil {
			t.Fatalf("AuthoritativeQuoteSources(%q) error: %v", value, err)
		}
	}
	for _, value := range []string{"", "BTCUSDT", "BA:BTCUSDT", "crypto::BTCUSDT"} {
		if err := AuthoritativeQuoteSources("source", value); err == nil {
			t.Fatalf("AuthoritativeQuoteSources(%q) unexpectedly succeeded", value)
		}
	}
}

func TestFundingRateSource(t *testing.T) {
	if err := FundingRateSource(int64(common.ContractType_CONTRACT_TYPE_PERPETUAL), "premium-v1"); err != nil {
		t.Fatalf("perpetual premium-v1 error: %v", err)
	}
	if err := FundingRateSource(int64(common.ContractType_CONTRACT_TYPE_PERPETUAL), "funding-v1"); err == nil {
		t.Fatal("perpetual funding-v1 unexpectedly succeeded")
	}
	if err := FundingRateSource(int64(common.ContractType_CONTRACT_TYPE_DELIVERY), ""); err != nil {
		t.Fatalf("delivery empty funding source error: %v", err)
	}
}

func TestContractPriceSources(t *testing.T) {
	if err := ContractPriceSources(
		int64(common.ContractType_CONTRACT_TYPE_PERPETUAL),
		"crypto:BA:BTCUSDT", "", 0, "",
	); err != nil {
		t.Fatalf("valid perpetual source rejected: %v", err)
	}
	if err := ContractPriceSources(
		int64(common.ContractType_CONTRACT_TYPE_DELIVERY),
		"", "crypto:BA:BTCUSD", 60, "delivery-btcusd-v1",
	); err != nil {
		t.Fatalf("valid delivery source rejected: %v", err)
	}

	tests := []struct {
		name       string
		source     string
		window     int64
		algorithm  string
		contract   common.ContractType
		markSource string
	}{
		{name: "bare perpetual source", contract: common.ContractType_CONTRACT_TYPE_PERPETUAL, markSource: "BTCUSDT"},
		{name: "bare delivery source", contract: common.ContractType_CONTRACT_TYPE_DELIVERY, source: "BTCUSD", window: 60, algorithm: "v1"},
		{name: "delivery without window", contract: common.ContractType_CONTRACT_TYPE_DELIVERY, source: "crypto:BA:BTCUSD", algorithm: "v1"},
		{name: "delivery without formula version", contract: common.ContractType_CONTRACT_TYPE_DELIVERY, source: "crypto:BA:BTCUSD", window: 60},
		{name: "non-contract product", contract: common.ContractType_CONTRACT_TYPE_NOT_APPLICABLE},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ContractPriceSources(
				int64(tt.contract), tt.markSource, tt.source, tt.window, tt.algorithm,
			); err == nil {
				t.Fatal("invalid contract price source configuration was accepted")
			}
		})
	}
}
