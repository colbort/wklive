package provider

import (
	"testing"

	"wklive/proto/liquidity"
	"wklive/proto/trade"
)

func TestNormalizeQuoteKeepsFreezingOrderUncertain(t *testing.T) {
	result, err := normalizeQuote(&trade.TradeOrder{
		Id: 12, OrderNo: "T12", Status: trade.OrderStatus_ORDER_STATUS_FREEZING, FilledQty: "0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN) {
		t.Fatalf("freezing order must remain unresolved, got status=%d", result.Status)
	}
	if result.LastErrorMsg == "" {
		t.Fatal("freezing order must carry a diagnostic reason")
	}
}

func TestNormalizeQuoteReportsUnknownTradeStatus(t *testing.T) {
	result, err := normalizeQuote(&trade.TradeOrder{
		Id: 13, OrderNo: "T13", Status: trade.OrderStatus(99), FilledQty: "0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN) || result.LastErrorMsg == "" {
		t.Fatalf("unknown trade status must be diagnostic and unresolved: %+v", result)
	}
}
