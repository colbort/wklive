package server

import "strings"

type ClientAction string
type Topic string

const (
	ActionSubscribe   ClientAction = "subscribe"
	ActionUnsubscribe ClientAction = "unsubscribe"

	TopicQuote Topic = "quote"
	TopicDepth Topic = "depth"
	TopicTick  Topic = "tick"
	TopicKline Topic = "kline"
)

type ClientMessage struct {
	Action   ClientAction `json:"action"`
	Topic    Topic        `json:"topic"`
	Market   string       `json:"market"`             // crypto
	Symbol   string       `json:"symbol"`             // BTCUSDT
	Region   string       `json:"region,omitempty"`   // BA
	Interval string       `json:"interval,omitempty"` // 1m/5m/15m/30m/1h/1d/1w/1mo
}

type ServerMessage struct {
	Topic    Topic  `json:"topic"`
	Market   string `json:"market"`
	Symbol   string `json:"symbol"`
	Region   string `json:"region,omitempty"`
	Interval string `json:"interval,omitempty"`
	Payload  any    `json:"payload"`
}

func BuildTopicKey(msg ClientMessage) string {
	return strings.ToLower(
		string(msg.Topic) + ":" +
			msg.Market + ":" +
			msg.Symbol + ":" +
			msg.Region + ":" +
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
		msg.Market = parts[1]
	}
	if len(parts) > 2 {
		msg.Symbol = parts[2]
	}
	if len(parts) > 3 {
		msg.Region = parts[3]
	}
	if len(parts) > 4 {
		msg.Interval = parts[4]
	}
	return msg
}
