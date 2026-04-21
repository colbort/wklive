package client

import (
	"testing"

	"wklive/services/itick/internal/socket/server"
)

func TestSubscriptionChangeToClientMessageNormalizesNonKlineInterval(t *testing.T) {
	msg := SubscriptionChange{
		Topic:        "quote",
		CategoryCode: " Crypto ",
		Symbol:       "btcusdt",
		Market:       "ba",
		Interval:     "1m",
	}.ToClientMessage()

	if msg.Interval != "" {
		t.Fatalf("expected quote interval to be empty, got %q", msg.Interval)
	}
	if msg.CategoryCode != "crypto" || msg.Symbol != "BTCUSDT" || msg.Market != "BA" {
		t.Fatalf("unexpected normalized message: %+v", msg)
	}
}

func TestRegistryTopicKeyIgnoresNonKlineInterval(t *testing.T) {
	withInterval := server.BuildTopicKey(SubscriptionChange{
		Topic:        "depth",
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "BA",
		Interval:     "1m",
	}.ToClientMessage())
	withoutInterval := server.BuildTopicKey(SubscriptionChange{
		Topic:        "depth",
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "BA",
	}.ToClientMessage())

	if withInterval != withoutInterval {
		t.Fatalf("expected depth topic keys to match, got %q and %q", withInterval, withoutInterval)
	}
}
