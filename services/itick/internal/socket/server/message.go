package server

import "strings"

type ClientAction string
type Topic string

const (
	TopicQuote Topic = "quote" // 实时价格
	TopicDepth Topic = "depth" // 盘口
	TopicTick  Topic = "tick"  // 实时成交
	TopicKline Topic = "kline" // kline
)

type ClientMessage struct {
	Topic        Topic  `json:"topic"`
	CategoryCode string `json:"categoryCode"`       // crypto
	Symbol       string `json:"symbol"`             // BTCUSDT
	Market       string `json:"market,omitempty"`   // BA
	Interval     string `json:"interval,omitempty"` // 1m/5m/15m/30m/1h/1d/1w/1mo
}

type ServerMessage struct {
	Topic        Topic  `json:"topic"`
	CategoryCode string `json:"categoryCode"`
	Symbol       string `json:"symbol"`
	Market       string `json:"market,omitempty"`
	Interval     string `json:"interval,omitempty"`
	Payload      any    `json:"payload"`
}

func BuildTopicKey(msg ClientMessage) string {
	msg = NormalizeClientMessage(msg)

	return strings.ToLower(
		string(msg.Topic) + ":" +
			msg.CategoryCode + ":" +
			msg.Symbol + ":" +
			msg.Market + ":" +
			msg.Interval,
	)
}

func ParseTopicKey(key string) ClientMessage {
	parts := strings.Split(key, ":")
	msg := ClientMessage{}
	if len(parts) > 0 {
		msg.Topic = Topic(parts[0])
	}
	if len(parts) > 1 {
		msg.CategoryCode = parts[1]
	}
	if len(parts) > 2 {
		msg.Symbol = parts[2]
	}
	if len(parts) > 3 {
		msg.Market = parts[3]
	}
	if len(parts) > 4 {
		msg.Interval = parts[4]
	}
	return NormalizeClientMessage(msg)
}

func NormalizeClientMessage(msg ClientMessage) ClientMessage {
	msg.Topic = Topic(strings.ToLower(strings.TrimSpace(string(msg.Topic))))
	msg.CategoryCode = strings.ToLower(strings.TrimSpace(msg.CategoryCode))
	msg.Symbol = strings.ToUpper(strings.TrimSpace(msg.Symbol))
	msg.Market = strings.ToUpper(strings.TrimSpace(msg.Market))
	msg.Interval = strings.ToLower(strings.TrimSpace(msg.Interval))

	if msg.Topic != TopicKline {
		msg.Interval = ""
	}

	return msg
}
