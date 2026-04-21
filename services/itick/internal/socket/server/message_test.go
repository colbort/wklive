package server

import "testing"

func TestNormalizeClientMessageClearsIntervalForNonKline(t *testing.T) {
	msg := NormalizeClientMessage(ClientMessage{
		Topic:        TopicQuote,
		CategoryCode: " Crypto ",
		Symbol:       "btcusdt",
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
		Symbol:       "btcusdt",
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
