package validation

import (
	"testing"

	"wklive/proto/trade"
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
	if err := FundingRateSource(int64(trade.ContractType_CONTRACT_TYPE_PERPETUAL), "premium-v1"); err != nil {
		t.Fatalf("perpetual premium-v1 error: %v", err)
	}
	if err := FundingRateSource(int64(trade.ContractType_CONTRACT_TYPE_PERPETUAL), "funding-v1"); err == nil {
		t.Fatal("perpetual funding-v1 unexpectedly succeeded")
	}
	if err := FundingRateSource(int64(trade.ContractType_CONTRACT_TYPE_DELIVERY), ""); err != nil {
		t.Fatalf("delivery empty funding source error: %v", err)
	}
}
