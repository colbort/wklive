package server

import "testing"

func TestNormalizeClientMessageClearsIntervalForNonKline(t *testing.T) {
	msg := NormalizeClientMessage(ClientMessage{
		Topic:        TopicQuote,
		CategoryCode: " Crypto ",
		Symbol:       "BTCUSDT",
		Market:       "ba",
		Interval:     "1m",
	})

	if msg.Interval != "" {
		t.Fatalf("expected non-kline interval to be empty, got %q", msg.Interval)
	}
	if msg.CategoryCode != "crypto" || msg.Symbol != "BTCUSDT" || msg.Market != "BA" {
		t.Fatalf("unexpected normalized message: %+v", msg)
	}
}

func TestNormalizeClientMessageKeepsKlineInterval(t *testing.T) {
	msg := NormalizeClientMessage(ClientMessage{
		Topic:        TopicKline,
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "ba",
		Interval:     "1M",
	})

	if msg.Interval != "1m" {
		t.Fatalf("expected kline interval to be kept and normalized, got %q", msg.Interval)
	}
}

func TestBuildTopicKeyIgnoresIntervalForQuote(t *testing.T) {
	withInterval := BuildTopicKey(ClientMessage{
		Topic:        TopicQuote,
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "BA",
		Interval:     "1m",
	})
	withoutInterval := BuildTopicKey(ClientMessage{
		Topic:        TopicQuote,
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "BA",
	})

	if withInterval != withoutInterval {
		t.Fatalf("expected quote topic keys to match, got %q and %q", withInterval, withoutInterval)
	}
}

func TestBuildTopicKeySeparatesTopicTypes(t *testing.T) {
	base := ClientMessage{
		CategoryCode: "crypto",
		Symbol:       "ETHUSDT",
		Market:       "BA",
	}

	quote := base
	quote.Topic = TopicQuote
	depth := base
	depth.Topic = TopicDepth
	tick := base
	tick.Topic = TopicTick
	kline := base
	kline.Topic = TopicKline
	kline.Interval = "1m"

	keys := map[string]struct{}{}
	for _, msg := range []ClientMessage{quote, depth, tick, kline} {
		key := BuildTopicKey(msg)
		if _, ok := keys[key]; ok {
			t.Fatalf("expected distinct topic key, got duplicate %q", key)
		}
		keys[key] = struct{}{}
	}
}
