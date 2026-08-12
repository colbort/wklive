// Package types exposes the provider-neutral market message and payload types
// used by the market service.
package types

import market "wklive/common/market"

type DepthLevel = market.DepthLevel
type DepthPayload = market.DepthPayload
type QuotePayload = market.QuotePayload
type TickPayload = market.TickPayload
type KlinePayload = market.KlinePayload
type Topic = market.Topic

const (
	TopicQuote = market.TopicQuote
	TopicDepth = market.TopicDepth
	TopicTick  = market.TopicTick
	TopicKline = market.TopicKline
)

type ClientMessage = market.ClientMessage
