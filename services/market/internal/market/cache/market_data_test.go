package cache

import (
	"testing"
	"wklive/services/market/internal/market/types"
)

func TestMarketDataKeyUsesLegacyKeyForRealtimeTopics(t *testing.T) {
	tests := []struct {
		topic types.Topic
		want  string
	}{
		{topic: types.TopicQuote, want: "market:quote:crypto:BA:BTCUSDT"},
		{topic: types.TopicDepth, want: "market:depth:crypto:BA:BTCUSDT"},
		{topic: types.TopicTick, want: "market:tick:crypto:BA:BTCUSDT"},
	}

	for _, tt := range tests {
		t.Run(string(tt.topic), func(t *testing.T) {
			got := marketDataKey(types.ClientMessage{
				Topic:        tt.topic,
				CategoryCode: " Crypto ",
				Market:       "ba",
				Symbol:       "btcusdt",
			})
			if got != tt.want {
				t.Fatalf("marketDataKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMarketDataKeyKeepsVersionedKlineKey(t *testing.T) {
	got := marketDataKey(types.ClientMessage{
		Topic:        types.TopicKline,
		CategoryCode: "crypto",
		Market:       "ba",
		Symbol:       "btcusdt",
		Interval:     "1M",
	})
	want := "market:v1:kline:crypto:BA:BTCUSDT:1m"
	if got != want {
		t.Fatalf("marketDataKey() = %q, want %q", got, want)
	}
}

func TestNormalizeClientMessageClearsIntervalForNonKline(t *testing.T) {
	msg := NormalizeClientMessage(types.ClientMessage{
		Topic:        types.TopicQuote,
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
	msg := NormalizeClientMessage(types.ClientMessage{
		Topic:        types.TopicKline,
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
	withInterval := BuildTopicKey(types.ClientMessage{
		Topic:        types.TopicQuote,
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "BA",
		Interval:     "1m",
	})
	withoutInterval := BuildTopicKey(types.ClientMessage{
		Topic:        types.TopicQuote,
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "BA",
	})

	if withInterval != withoutInterval {
		t.Fatalf("expected quote topic keys to match, got %q and %q", withInterval, withoutInterval)
	}
}

func TestBuildTopicKeySeparatesTopicTypes(t *testing.T) {
	base := types.ClientMessage{
		CategoryCode: "crypto",
		Symbol:       "ETHUSDT",
		Market:       "BA",
	}

	quote := base
	quote.Topic = types.TopicQuote
	depth := base
	depth.Topic = types.TopicDepth
	tick := base
	tick.Topic = types.TopicTick
	kline := base
	kline.Topic = types.TopicKline
	kline.Interval = "1m"

	keys := map[string]struct{}{}
	for _, msg := range []types.ClientMessage{quote, depth, tick, kline} {
		key := BuildTopicKey(msg)
		if _, ok := keys[key]; ok {
			t.Fatalf("expected distinct topic key, got duplicate %q", key)
		}
		keys[key] = struct{}{}
	}
}
