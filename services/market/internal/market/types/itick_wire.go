package types

import (
	"encoding/json"
	"strings"

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
	S      string          `json:"s,omitempty"`
	R      string          `json:"r,omitempty"`
	Type   string          `json:"type,omitempty"`
	LD     float64         `json:"ld,omitempty"`
	O      float64         `json:"o,omitempty"`
	H      float64         `json:"h,omitempty"`
	L      float64         `json:"l,omitempty"`
	C      float64         `json:"c,omitempty"`
	V      float64         `json:"v,omitempty"`
	TU     float64         `json:"tu,omitempty"`
	T      int64           `json:"t,omitempty"`
	A      json.RawMessage `json:"a,omitempty"`
	B      json.RawMessage `json:"b,omitempty"`
	LDText string          `json:"-"`
}

// UnmarshalJSON preserves the exact upstream last-price token while retaining
// float fields for non-settlement calculations.
func (d *UpstreamData) UnmarshalJSON(data []byte) error {
	type wire UpstreamData
	var decoded wire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*d = UpstreamData(decoded)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	raw := strings.TrimSpace(string(fields["ld"]))
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	d.LDText = raw
	return nil
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
