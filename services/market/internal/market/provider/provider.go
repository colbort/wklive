// Package provider defines the vendor-neutral realtime market-data boundary.
// A new upstream integration implements these interfaces and can then be used
// by MarketManager without leaking its REST or WebSocket protocol details.
package provider

import (
	"context"

	market "wklive/common/market"
)

type Subscription = market.ClientMessage
type Quote = market.QuotePayload

// Stream owns one provider stream for a market category. Implementations may
// use WebSocket, SSE, polling, or another transport while preserving the same
// subscription lifecycle for MarketManager.
type Stream interface {
	Start(context.Context)
	HasDesiredSubscriptions() bool
	ReplaceSubscriptions([]Subscription) error
	Resubscribe(Subscription) error
	IsLeader() bool
	SetReconnectHandler(func(category string))
}

// RealtimeProvider is the complete capability required by the live market
// pipeline. Provider implementations normalize their native data into the
// common market payloads before writing to the shared market cache.
type RealtimeProvider interface {
	Code() string
	// Categories returns the canonical business categories for which this
	// provider can create streams. Stream discovery must not depend on another
	// vendor's category table.
	Categories() []string
	Supports(category string) bool
	Warm(context.Context, []Subscription)
	FetchQuote(context.Context, Subscription) (*Quote, error)
	NewStream(category string) (Stream, error)
}

// ProviderGroup exposes the independently managed providers inside an
// aggregate quote source. MarketManager uses this to own one stream per
// provider and category while the aggregate still handles REST selection.
type ProviderGroup interface {
	Providers() []RealtimeProvider
}

func Sources(source RealtimeProvider) []RealtimeProvider {
	if source == nil {
		return nil
	}
	if group, ok := source.(ProviderGroup); ok {
		return group.Providers()
	}
	return []RealtimeProvider{source}
}
