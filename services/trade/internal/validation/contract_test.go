package validation

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"
)

func TestValidateSymbolTradingTimeline(t *testing.T) {
	tests := []struct {
		name    string
		listing int64
		start   int64
		end     int64
		wantErr bool
	}{
		{name: "valid", listing: 100, start: 200, end: 300},
		{name: "start equals listing", listing: 100, start: 100, end: 300, wantErr: true},
		{name: "end before start", listing: 100, start: 300, end: 200, wantErr: true},
		{name: "delivery requires all times", start: 200, end: 300, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SymbolTradingTimeline(trade.ProductType_PRODUCT_TYPE_DERIVATIVE, trade.ContractType_CONTRACT_TYPE_DELIVERY, tt.listing, tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateSymbolTradingTimeline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateContractTradingTimeline(t *testing.T) {
	symbol := &models.TTradeSymbol{
		ProductType:      int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE),
		ContractType:     int64(trade.ContractType_CONTRACT_TYPE_DELIVERY),
		ListingTime:      100,
		TradingStartTime: 200,
		TradingEndTime:   500,
	}
	if err := ContractTradingTimeline(symbol, 600, 300, 500); err != nil {
		t.Fatalf("valid delivery timeline rejected: %v", err)
	}
	if err := ContractTradingTimeline(symbol, 500, 300, 400); err == nil {
		t.Fatal("delivery at trading end should be rejected")
	}
	if err := ContractTradingTimeline(symbol, 600, 300, 501); err == nil {
		t.Fatal("matching after trading end should be rejected")
	}
}

func TestValidateContractMarginModes(t *testing.T) {
	if err := ContractMarginModes(0, 1); err != nil {
		t.Fatalf("isolated-only mode rejected: %v", err)
	}
	if err := ContractMarginModes(0, 0); err == nil {
		t.Fatal("configuration without a supported margin mode should be rejected")
	}
	if err := ContractMarginModes(2, 1); err == nil {
		t.Fatal("invalid support flag should be rejected")
	}
}

func TestContractSupportsMarginMode(t *testing.T) {
	config := &models.TTradeSymbolContract{SupportCross: 0, SupportIsolated: 1}
	if ContractSupportsMarginMode(config, trade.MarginMode_MARGIN_MODE_CROSS) {
		t.Fatal("unsupported cross margin was accepted")
	}
	if !ContractSupportsMarginMode(config, trade.MarginMode_MARGIN_MODE_ISOLATED) {
		t.Fatal("supported isolated margin was rejected")
	}
	if ContractSupportsMarginMode(config, trade.MarginMode_MARGIN_MODE_UNKNOWN) {
		t.Fatal("unknown margin mode was accepted")
	}
}
