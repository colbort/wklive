package tasklogic

import (
	"context"
	"testing"

	"wklive/proto/common"
	"wklive/proto/liquidity"
	"wklive/proto/market"
	"wklive/services/liquidity/internal/provider"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type marketClientStub struct {
	statusReq    *market.GetTradingStatusReq
	statusResp   *market.GetTradingStatusResp
	statusErr    error
	snapshotResp *market.GetAuthoritativeSnapshotResp
	snapshotErr  error
}

type quoteOrderModelStub struct {
	models.TLiquidityQuoteOrderModel
	rows    []*models.TLiquidityQuoteOrder
	updated []*models.TLiquidityQuoteOrder
}

type providerModelStub struct {
	models.TLiquidityProviderModel
	row *models.TLiquidityProvider
}

func (s *providerModelStub) FindOne(context.Context, int64) (*models.TLiquidityProvider, error) {
	return s.row, nil
}

type internalMarketMakerStub struct {
	cancelResult *provider.QuoteResult
	cancelErr    error
}

func (s *internalMarketMakerStub) Health(context.Context, *models.TLiquidityProvider) error {
	return nil
}

func (s *internalMarketMakerStub) PlaceQuote(context.Context, *models.TLiquidityProvider, *models.TLiquidityQuoteOrder) (*provider.QuoteResult, error) {
	return nil, nil
}

func (s *internalMarketMakerStub) CancelQuote(context.Context, *models.TLiquidityProvider, *models.TLiquidityQuoteOrder) (*provider.QuoteResult, error) {
	return s.cancelResult, s.cancelErr
}

func (s *internalMarketMakerStub) QueryQuote(context.Context, *models.TLiquidityProvider, *models.TLiquidityQuoteOrder) (*provider.QuoteResult, error) {
	return nil, nil
}

func (s *quoteOrderModelStub) FindActiveByConfig(context.Context, int64) ([]*models.TLiquidityQuoteOrder, error) {
	return s.rows, nil
}

func (s *quoteOrderModelStub) Update(_ context.Context, row *models.TLiquidityQuoteOrder) error {
	s.updated = append(s.updated, row)
	return nil
}

func (s *marketClientStub) GetAuthoritativeSnapshot(context.Context, *market.GetAuthoritativeSnapshotReq, ...grpc.CallOption) (*market.GetAuthoritativeSnapshotResp, error) {
	return s.snapshotResp, s.snapshotErr
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

func TestEnsureMarketOpenCancelsPendingQuotesWhenClosed(t *testing.T) {
	client := &marketClientStub{statusResp: &market.GetTradingStatusResp{
		Base: &common.RespBase{Code: 200},
		Data: &market.GetTradingStatusData{IsOpen: false, Reason: "market_closed"},
	}}
	row := &models.TLiquidityQuoteOrder{
		Id: 77, ConfigId: 7,
		Status: int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT),
	}
	orders := &quoteOrderModelStub{rows: []*models.TLiquidityQuoteOrder{row}}
	config := &models.TLiquiditySymbolConfig{
		Id: 7, Symbol: "AAPL", ReferencePriceSource: "stock:US:AAPL",
	}
	open, err := ensureMarketOpen(context.Background(), &svc.ServiceContext{
		MarketClient: client, QuoteOrderModel: orders,
	}, config, 12345)
	if err != nil {
		t.Fatal(err)
	}
	if open {
		t.Fatal("closed market must not allow liquidity quoting")
	}
	if len(orders.updated) != 1 || row.Status != int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED) || row.CancelReason != "market_closed" {
		t.Fatalf("pending quote was not canceled: %+v", row)
	}
}

func TestLoadReferenceQuoteOrCancelCancelsPendingQuotesWhenReferenceUnavailable(t *testing.T) {
	row := &models.TLiquidityQuoteOrder{
		Id: 78, ConfigId: 8,
		Status: int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PENDING_SUBMIT),
	}
	orders := &quoteOrderModelStub{rows: []*models.TLiquidityQuoteOrder{row}}
	config := &models.TLiquiditySymbolConfig{
		Id: 8, Symbol: "USDCNY", ReferencePriceSource: "forex:GB:USDCNY",
	}
	_, err := loadReferenceQuoteOrCancel(context.Background(), &svc.ServiceContext{
		MarketClient: &marketClientStub{}, QuoteOrderModel: orders,
	}, config, 12345)
	if err == nil {
		t.Fatal("missing reference price must fail")
	}
	if len(orders.updated) != 1 || row.Status != int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED) || row.CancelReason != "reference price unavailable" {
		t.Fatalf("pending quote was not canceled after reference price failure: %+v", row)
	}
}

func TestCancelActiveQuotesPreservesSystemReason(t *testing.T) {
	row := &models.TLiquidityQuoteOrder{
		Id: 79, ConfigId: 9, ProviderId: 3,
		Status: int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_OPEN),
	}
	orders := &quoteOrderModelStub{rows: []*models.TLiquidityQuoteOrder{row}}
	err := cancelActiveQuotes(context.Background(), &svc.ServiceContext{
		QuoteOrderModel: orders,
		ProviderModel:   &providerModelStub{row: &models.TLiquidityProvider{Id: 3}},
		InternalMarketMaker: &internalMarketMakerStub{cancelResult: &provider.QuoteResult{
			Status: int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED),
			Reason: "canceled by user",
		}},
	}, 9, "market_closed")
	if err != nil {
		t.Fatal(err)
	}
	if len(orders.updated) != 1 || row.Status != int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED) || row.CancelReason != "market_closed" {
		t.Fatalf("system cancellation reason was not preserved: %+v", row)
	}
}
