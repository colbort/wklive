package client

import "encoding/json"

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

type UpstreamPongData struct {
	Params string `json:"params"`
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

// 深度档位
type DepthLevel struct {
	Price        float64 `json:"p"`
	Volume       float64 `json:"v"`
	Position     int64   `json:"po"`
	OriginVolume float64 `json:"o"`
}

// 深度
type DepthPayload struct {
	Asks []*DepthLevel
	Bids []*DepthLevel
}

type QuotePayload struct {
	LastPrice float64 `json:"lastPrice"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Volume    float64 `json:"volume"`
	Turnover  float64 `json:"turnover"`
	Ts        int64   `json:"ts"`
}

type TickPayload struct {
	LastPrice float64 `json:"lastPrice"`
	Volume    float64 `json:"volume"`
	Ts        int64   `json:"ts"`
}

type KlinePayload struct {
	Interval string  `json:"interval"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Volume   float64 `json:"volume"`
	Turnover float64 `json:"turnover"`
	Ts       int64   `json:"ts"`
}
