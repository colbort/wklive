package types

import (
	"encoding/json"
	cache "wklive/common/market"
)

type SubscribeReq struct {
	Ac     string `json:"ac"`
	Params string `json:"params"`
	Types  string `json:"types"`
}
type UnsubscribeReq struct {
	Ac     string `json:"ac"`
	Params string `json:"params"`
	Types  string `json:"types"`
}
type PingReq struct {
	Ac     string `json:"ac"`
	Params string `json:"params"`
}

type UpstreamEnvelope struct {
	Code  int             `json:"code"`
	ResAc string          `json:"resAc,omitempty"`
	Msg   string          `json:"msg,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}
type UpstreamData struct {
	S    string          `json:"s,omitempty"`
	R    string          `json:"r,omitempty"`
	Type string          `json:"type,omitempty"`
	LD   float64         `json:"ld,omitempty"`
	O    float64         `json:"o,omitempty"`
	H    float64         `json:"h,omitempty"`
	L    float64         `json:"l,omitempty"`
	C    float64         `json:"c,omitempty"`
	V    float64         `json:"v,omitempty"`
	TU   float64         `json:"tu,omitempty"`
	T    int64           `json:"t,omitempty"`
	A    json.RawMessage `json:"a,omitempty"`
	B    json.RawMessage `json:"b,omitempty"`
}
type DepthLevel = cache.DepthLevel
type DepthPayload = cache.DepthPayload
type QuotePayload = cache.QuotePayload
type TickPayload = cache.TickPayload
type KlinePayload = cache.KlinePayload
type Topic = cache.Topic

const (
	TopicQuote = cache.TopicQuote
	TopicDepth = cache.TopicDepth
	TopicTick  = cache.TopicTick
	TopicKline = cache.TopicKline
)

type ClientMessage = cache.ClientMessage
