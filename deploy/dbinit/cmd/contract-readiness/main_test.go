package main

import (
	"reflect"
	"testing"
)

func TestLoadInputDetailed(t *testing.T) {
	t.Setenv("READINESS_SOURCE_AUTHORITIES", " source-a,source-b,source-a,source-c ")
	t.Setenv("READINESS_SOURCE_MARKETS", " market-a,market-b,market-c ")
	t.Setenv("READINESS_INDEX_SOURCE_WEIGHTS", " 1,2,3 ")
	t.Setenv("READINESS_SOURCE_WEIGHTS", " 1,2,3 ")
	t.Setenv("READINESS_CATEGORY_CODE", "crypto")
	t.Setenv("READINESS_MARKET", "BA")
	t.Setenv("READINESS_PRICE_SYMBOL", "BTCUSDT")
	t.Setenv("READINESS_PERPETUAL_SYMBOL", "BTCUSDT-PERP")
	t.Setenv("READINESS_DELIVERY_SYMBOL", "BTCUSDT-20260925")
	t.Setenv("READINESS_PERPETUAL_PRICE_AUTHORITY", "perpetual-source")
	t.Setenv("READINESS_PERPETUAL_PRICE_MARKET", "PERPETUAL")
	t.Setenv("READINESS_TENANT_ID", "900101")
	t.Setenv("READINESS_SETTLEMENT_COIN", "USDT")
	t.Setenv("READINESS_INSURANCE_FUND_MIN_AVAILABLE", "100000")
	t.Setenv("READINESS_INDEX_ALGORITHM", "2")
	t.Setenv("READINESS_INDEX_FORMULA_VERSION", "index-v1")
	t.Setenv("READINESS_INDEX_MAX_DEVIATION_BPS", "200")
	t.Setenv("READINESS_MARK_FORMULA_VERSION", "mark-v2")
	t.Setenv("READINESS_MARK_MAX_BASIS_BPS", "200")
	t.Setenv("READINESS_MARK_CURRENT_WEIGHT", "1")
	t.Setenv("READINESS_MARK_PREVIOUS_WEIGHT", "4")
	t.Setenv("READINESS_FUNDING_FORMULA_VERSION", "funding-v1")
	t.Setenv("READINESS_FORMULA_INTERVAL_MS", "1000")
	t.Setenv("READINESS_DELIVERY_ALGORITHM", "2")
	t.Setenv("READINESS_FORMULA_VERSION", "delivery-v1")
	t.Setenv("READINESS_MAX_LOOKBACK_MS", "30000")
	t.Setenv("READINESS_MAX_DEVIATION_BPS", "200")

	input, err := loadInput(true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input.SourceAuthorities, []string{"source-a", "source-b", "source-c"}) {
		t.Fatalf("unexpected sources: %#v", input.SourceAuthorities)
	}
	if !reflect.DeepEqual(input.DeliverySourceWeights, []string{"1", "2", "3"}) {
		t.Fatalf("unexpected weights: %#v", input.DeliverySourceWeights)
	}
	if !reflect.DeepEqual(input.SourceMarkets, []string{"market-a", "market-b", "market-c"}) ||
		!reflect.DeepEqual(input.IndexSourceWeights, []string{"1", "2", "3"}) {
		t.Fatalf("unexpected source dimensions: %+v", input)
	}
	if input.TenantID != 900101 ||
		input.InsuranceFundMinAvailable != "100000" ||
		input.IndexAlgorithm != 2 ||
		input.DeliveryAlgorithm != 2 ||
		input.IndexMaxDeviationBps != 200 ||
		input.MarkMaxBasisBps != 200 ||
		input.PriceFormulaIntervalMs != 1000 ||
		input.DeliveryMaxLookbackMs != 30000 ||
		input.DeliveryMaxDeviationBps != 200 {
		t.Fatalf("unexpected input: %+v", input)
	}
}

func TestLoadInputOpenOnlyDoesNotRequireDimensions(t *testing.T) {
	input, err := loadInput(false)
	if err != nil {
		t.Fatal(err)
	}
	if input.TenantID != 0 || len(input.SourceAuthorities) != 0 {
		t.Fatalf("unexpected input: %+v", input)
	}
}

func TestParsePositiveInt64RejectsInvalidValue(t *testing.T) {
	t.Setenv("INVALID_POSITIVE", "0")
	if _, err := parsePositiveInt64("INVALID_POSITIVE"); err == nil {
		t.Fatal("expected validation error")
	}
}
