package tasklogic

import (
	"context"
	"testing"

	"wklive/proto/common"
	"wklive/proto/market"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type marketClientStub struct {
	statusReq  *market.GetTradingStatusReq
	statusResp *market.GetTradingStatusResp
	statusErr  error
}

func (s *marketClientStub) GetAuthoritativeSnapshot(context.Context, *market.GetAuthoritativeSnapshotReq, ...grpc.CallOption) (*market.GetAuthoritativeSnapshotResp, error) {
	return nil, nil
}

func (s *marketClientStub) GetTradingStatus(_ context.Context, in *market.GetTradingStatusReq, _ ...grpc.CallOption) (*market.GetTradingStatusResp, error) {
	s.statusReq = in
	return s.statusResp, s.statusErr
}

func TestBuildQuoteOrders(t *testing.T) {
	config := &models.TLiquiditySymbolConfig{
		Id: 7, SymbolId: 6, InternalProviderId: 3,
		BaseSpreadBps: decimal.NewFromInt(5), MaxSpreadBps: decimal.NewFromInt(50), MaxPriceDeviationBps: decimal.NewFromInt(40),
		PriceTick: decimal.RequireFromString("0.1"), QtyStep: decimal.RequireFromString("0.001"), MinQuoteQty: decimal.RequireFromString("0.01"),
		MaxQuoteQty: decimal.NewFromInt(2), MaxQuoteNotional: decimal.NewFromInt(150), QuoteTtlMs: 5_000,
	}
	levels := []*models.TLiquidityStrategyLevel{{
		LevelNo: 1, BidSpreadBps: decimal.NewFromInt(5), AskSpreadBps: decimal.NewFromInt(10),
		BidQty: decimal.NewFromInt(3), AskQty: decimal.RequireFromString("1.2345"), Enabled: int64(common.Switch_SWITCH_ON),
	}}
	orders := buildQuoteOrders(config, levels, decimal.NewFromInt(100), 1_000)
	if len(orders) != 2 {
		t.Fatalf("expected two-sided quotes, got %d", len(orders))
	}
	if orders[0].Side != int64(common.Side_SIDE_BUY) || !orders[0].Price.Equal(decimal.RequireFromString("99.9")) || !orders[0].Qty.Equal(decimal.RequireFromString("1.501")) {
		t.Fatalf("unexpected bid: %+v", orders[0])
	}
	if orders[1].Side != int64(common.Side_SIDE_SELL) || !orders[1].Price.Equal(decimal.RequireFromString("100.2")) || !orders[1].Qty.Equal(decimal.RequireFromString("1.234")) {
		t.Fatalf("unexpected ask: %+v", orders[1])
	}
	if orders[0].ExpireAt != 6_000 || orders[1].ExpireAt != 6_000 {
		t.Fatalf("unexpected expiry")
	}
}

func TestParseReferenceSource(t *testing.T) {
	category, market, symbol := parseReferenceSource("crypto:BA:BTCUSDT", "fallback")
	if category != "crypto" || market != "BA" || symbol != "BTCUSDT" {
		t.Fatalf("unexpected parsed source: %s/%s/%s", category, market, symbol)
	}
	_, _, symbol = parseReferenceSource("crypto:BA", "BTCUSDT")
	if symbol != "BTCUSDT" {
		t.Fatalf("expected fallback symbol, got %s", symbol)
	}
}

func TestStepRoundingRemovesFloatTail(t *testing.T) {
	ask := roundUp(decimal.RequireFromString("64088.02399999999"), decimal.RequireFromString("0.001"))
	if got := ask.String(); got != "64088.024" {
		t.Fatalf("unexpected normalized ask price: %s", got)
	}
	bid := roundDown(decimal.RequireFromString("63956.139800000004"), decimal.RequireFromString("0.0001"))
	if got := bid.String(); got != "63956.1398" {
		t.Fatalf("unexpected normalized bid price: %s", got)
	}
}

func TestLoadTradingStatusUsesPrimaryReferenceSource(t *testing.T) {
	client := &marketClientStub{statusResp: &market.GetTradingStatusResp{
		Base: &common.RespBase{Code: 200},
		Data: &market.GetTradingStatusData{IsOpen: true, Reason: "market_open"},
	}}
	config := &models.TLiquiditySymbolConfig{
		Symbol:               "AAPL",
		ReferencePriceSource: "stock:US:AAPL,crypto:BA:BTCUSDT",
	}
	open, reason, err := loadTradingStatus(context.Background(), &svc.ServiceContext{MarketClient: client}, config, 12345)
	if err != nil {
		t.Fatal(err)
	}
	if !open || reason != "market_open" {
		t.Fatalf("unexpected trading status: open=%v reason=%s", open, reason)
	}
	if client.statusReq.GetCategoryCode() != "stock" || client.statusReq.GetMarket() != "US" || client.statusReq.GetSymbol() != "AAPL" || client.statusReq.GetTimestamp() != 12345 {
		t.Fatalf("unexpected trading status request: %+v", client.statusReq)
	}
}
