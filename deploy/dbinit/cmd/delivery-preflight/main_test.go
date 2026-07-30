package main

import "testing"

func TestLoadInput(t *testing.T) {
	t.Setenv("DELIVERY_PREFLIGHT_TENANT_ID", "1")
	t.Setenv("DELIVERY_PREFLIGHT_SYMBOL", " BTCUSDT-20260925 ")
	t.Setenv("DELIVERY_PREFLIGHT_SETTLEMENT_ASSET", " USDT ")
	t.Setenv("DELIVERY_PREFLIGHT_CATEGORY_CODE", " crypto ")
	t.Setenv("DELIVERY_PREFLIGHT_MARKET", " BA ")
	t.Setenv("DELIVERY_PREFLIGHT_PRICE_SYMBOL", " BTCUSDT ")
	t.Setenv("DELIVERY_PREFLIGHT_FORMULA_VERSION", " delivery-v1 ")
	t.Setenv("DELIVERY_PREFLIGHT_FORMULA_ALGORITHM", "2")
	t.Setenv("DELIVERY_PREFLIGHT_MAX_LOOKBACK_MS", "30000")
	t.Setenv("DELIVERY_PREFLIGHT_MAX_DEVIATION_BPS", "200")
	t.Setenv("DELIVERY_PREFLIGHT_MIN_INPUT_COUNT", "3")
	t.Setenv("DELIVERY_PREFLIGHT_INTERVAL_MS", "1000")

	input, err := loadInput()
	if err != nil {
		t.Fatal(err)
	}
	if input.TenantID != 1 ||
		input.Symbol != "BTCUSDT-20260925" ||
		input.SettlementAsset != "USDT" ||
		input.CategoryCode != "crypto" ||
		input.Market != "BA" ||
		input.PriceSymbol != "BTCUSDT" ||
		input.FormulaVersion != "delivery-v1" ||
		input.FormulaAlgorithm != 2 ||
		input.FormulaMaxLookbackMs != 30000 ||
		input.FormulaMaxDeviationBps != 200 ||
		input.FormulaMinInputCount != 3 ||
		input.FormulaIntervalMs != 1000 {
		t.Fatalf("unexpected input: %+v", input)
	}
}

func TestLoadInputRejectsInvalidNumber(t *testing.T) {
	t.Setenv("DELIVERY_PREFLIGHT_TENANT_ID", "0")
	if _, err := loadInput(); err == nil {
		t.Fatal("expected validation error")
	}
}
