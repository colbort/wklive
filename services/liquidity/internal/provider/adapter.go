package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"wklive/services/liquidity/models"

	"github.com/shopspring/decimal"
)

// OrderResult is the normalized result returned by an external venue adapter.
// RawResponse must not contain credentials or other secrets.
type OrderResult struct {
	ExternalOrderID string
	Status          int64
	FilledQty       decimal.Decimal
	AvgPrice        decimal.Decimal
	FeeAmount       decimal.Decimal
	FeeAsset        string
	RawResponse     string
}

type Fill struct {
	ExternalTradeID string
	Side            int64
	Price           decimal.Decimal
	Qty             decimal.Decimal
	Amount          decimal.Decimal
	FeeAmount       decimal.Decimal
	FeeAsset        string
	LiquidityType   int64
	TradeTime       int64
	RawPayload      string
}

type Inventory struct {
	BaseAsset      string
	QuoteAsset     string
	BaseTotal      decimal.Decimal
	BaseAvailable  decimal.Decimal
	BaseFrozen     decimal.Decimal
	QuoteTotal     decimal.Decimal
	QuoteAvailable decimal.Decimal
	QuoteFrozen    decimal.Decimal
	PositionQty    decimal.Decimal
	RawPayload     string
}

// Adapter isolates exchange-specific authentication and request formats from
// the liquidity domain. Implementations are registered by venue code.
type Adapter interface {
	Health(ctx context.Context, provider *models.TLiquidityProvider) error
	SubmitOrder(ctx context.Context, provider *models.TLiquidityProvider, order *models.TLiquidityExternalOrder) (*OrderResult, error)
	CancelOrder(ctx context.Context, provider *models.TLiquidityProvider, order *models.TLiquidityExternalOrder) (*OrderResult, error)
	QueryOrder(ctx context.Context, provider *models.TLiquidityProvider, order *models.TLiquidityExternalOrder) (*OrderResult, error)
	QueryFills(ctx context.Context, provider *models.TLiquidityProvider, order *models.TLiquidityExternalOrder) ([]Fill, error)
	SnapshotInventory(ctx context.Context, provider *models.TLiquidityProvider, config *models.TLiquiditySymbolConfig) (*Inventory, error)
}

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[string]Adapter)}
}

func (r *Registry) Register(venueCode string, adapter Adapter) error {
	code := strings.ToUpper(strings.TrimSpace(venueCode))
	if code == "" || adapter == nil {
		return fmt.Errorf("venue code and adapter are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.adapters[code]; exists {
		return fmt.Errorf("adapter already registered for venue %s", code)
	}
	r.adapters[code] = adapter
	return nil
}

func (r *Registry) Get(venueCode string) (Adapter, error) {
	if r == nil {
		return nil, fmt.Errorf("provider adapter registry is nil")
	}
	code := strings.ToUpper(strings.TrimSpace(venueCode))
	r.mu.RLock()
	adapter := r.adapters[code]
	r.mu.RUnlock()
	if adapter == nil {
		return nil, fmt.Errorf("provider adapter is not configured for venue %s", code)
	}
	return adapter, nil
}
