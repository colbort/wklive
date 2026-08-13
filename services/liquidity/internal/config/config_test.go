package config

import "testing"

func TestValidateRequiresMarketAuthorities(t *testing.T) {
	if err := (Config{}).Validate(); err == nil {
		t.Fatal("empty MarketAuthorities must be rejected")
	}
	if err := (Config{MarketAuthorities: []string{"  "}}).Validate(); err == nil {
		t.Fatal("blank MarketAuthorities must be rejected")
	}
	if err := (Config{MarketAuthorities: []string{"itick-ws"}}).Validate(); err != nil {
		t.Fatalf("configured MarketAuthorities must be accepted: %v", err)
	}
}
