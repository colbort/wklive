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
