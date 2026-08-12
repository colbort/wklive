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
	Supports(category string) bool
	Warm(context.Context, []Subscription)
	FetchQuote(context.Context, Subscription) (*Quote, error)
	NewStream(category string) (Stream, error)
}
