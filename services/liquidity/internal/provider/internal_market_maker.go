package provider

import (
	"context"
	"fmt"
	"strconv"

	"wklive/proto/common"
	"wklive/proto/liquidity"
	"wklive/proto/trade"
	"wklive/services/liquidity/models"
)

type QuoteResult struct {
	InternalOrderID int64
	OrderNo         string
	Status          int64
	FilledQty       float64
	Reason          string
}

type InternalMarketMaker interface {
	Health(context.Context, *models.TLiquidityProvider) error
	PlaceQuote(context.Context, *models.TLiquidityProvider, *models.TLiquidityQuoteOrder) (*QuoteResult, error)
	CancelQuote(context.Context, *models.TLiquidityProvider, *models.TLiquidityQuoteOrder) (*QuoteResult, error)
	QueryQuote(context.Context, *models.TLiquidityProvider, *models.TLiquidityQuoteOrder) (*QuoteResult, error)
}

type TradeInternalMarketMaker struct{ client trade.TradeClient }

func NewTradeInternalMarketMaker(client trade.TradeClient) *TradeInternalMarketMaker {
	return &TradeInternalMarketMaker{client: client}
}

func (m *TradeInternalMarketMaker) Health(_ context.Context, p *models.TLiquidityProvider) error {
	if m == nil || m.client == nil {
		return fmt.Errorf("trade client is not configured")
	}
	if p == nil || p.TradeUserId <= 0 || p.ProviderType != int64(liquidity.ProviderType_PROVIDER_TYPE_INTERNAL) {
		return fmt.Errorf("valid internal liquidity provider is required")
	}
	return nil
}

func (m *TradeInternalMarketMaker) PlaceQuote(ctx context.Context, p *models.TLiquidityProvider, q *models.TLiquidityQuoteOrder) (*QuoteResult, error) {
	if err := m.Health(ctx, p); err != nil {
		return nil, err
	}
	if q == nil || q.SymbolId <= 0 || q.Price <= 0 || q.Qty <= 0 {
		return nil, fmt.Errorf("valid quote order is required")
	}
	resp, err := m.client.PlaceLiquidityQuote(ctx, &trade.PlaceLiquidityQuoteReq{TradeUserId: p.TradeUserId, Order: &trade.PlaceOrderReq{
		SymbolId: q.SymbolId, Side: common.Side(q.Side), OrderType: trade.OrderType_ORDER_TYPE_LIMIT,
		TimeInForce: trade.TimeInForce_TIME_IN_FORCE_GTC, ClientOrderId: q.ClientOrderId,
		Price: strconv.FormatFloat(q.Price, 'f', -1, 64), Qty: strconv.FormatFloat(q.Qty, 'f', -1, 64),
		OrderSource: trade.OrderSourceType_ORDER_SOURCE_TYPE_SYSTEM,
	}})
	if err != nil {
		return nil, err
	}
	if resp.GetBase().GetCode() != 200 || resp.Data == nil {
		return nil, fmt.Errorf("place liquidity quote failed: %s", resp.GetBase().GetMsg())
	}
	return normalizeQuote(resp.Data), nil
}

func (m *TradeInternalMarketMaker) CancelQuote(ctx context.Context, p *models.TLiquidityProvider, q *models.TLiquidityQuoteOrder) (*QuoteResult, error) {
	if err := m.Health(ctx, p); err != nil {
		return nil, err
	}
	resp, err := m.client.CancelLiquidityQuote(ctx, &trade.CancelLiquidityQuoteReq{TradeUserId: p.TradeUserId, Order: &trade.CancelOrderReq{OrderId: q.InternalOrderId, OrderNo: q.InternalOrderNo, ClientOrderId: q.ClientOrderId}})
	if err != nil {
		return nil, err
	}
	if resp.GetBase().GetCode() != 200 {
		return nil, fmt.Errorf("cancel liquidity quote failed: %s", resp.GetBase().GetMsg())
	}
	return m.QueryQuote(ctx, p, q)
}

func (m *TradeInternalMarketMaker) QueryQuote(ctx context.Context, p *models.TLiquidityProvider, q *models.TLiquidityQuoteOrder) (*QuoteResult, error) {
	if err := m.Health(ctx, p); err != nil {
		return nil, err
	}
	resp, err := m.client.GetLiquidityQuote(ctx, &trade.GetLiquidityQuoteReq{TradeUserId: p.TradeUserId, Order: &trade.GetOrderDetailReq{OrderId: q.InternalOrderId, OrderNo: q.InternalOrderNo}})
	if err != nil {
		return nil, err
	}
	if resp.GetBase().GetCode() != 200 || resp.GetData().GetOrder() == nil {
		return nil, fmt.Errorf("query liquidity quote failed: %s", resp.GetBase().GetMsg())
	}
	return normalizeQuote(resp.Data.Order), nil
}

func normalizeQuote(o *trade.TradeOrder) *QuoteResult {
	filled, _ := strconv.ParseFloat(o.FilledQty, 64)
	status := liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN
	switch o.Status {
	case trade.OrderStatus_ORDER_STATUS_PENDING:
		status = liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_OPEN
	case trade.OrderStatus_ORDER_STATUS_PART_FILLED:
		status = liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PART_FILLED
	case trade.OrderStatus_ORDER_STATUS_FILLED:
		status = liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FILLED
	case trade.OrderStatus_ORDER_STATUS_CANCELING:
		status = liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELING
	case trade.OrderStatus_ORDER_STATUS_CANCELED, trade.OrderStatus_ORDER_STATUS_EXPIRED:
		status = liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED
	case trade.OrderStatus_ORDER_STATUS_REJECTED:
		status = liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FAILED
	}
	return &QuoteResult{InternalOrderID: o.Id, OrderNo: o.OrderNo, Status: int64(status), FilledQty: filled, Reason: o.CancelReason}
}
