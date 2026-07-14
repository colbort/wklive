package client

import (
	"testing"
	"wklive/services/itick/internal/socket/cache"
	"wklive/services/itick/internal/socket/types"
)

func TestEnsureDesiredSubscriptionsMergesWithoutReplacing(t *testing.T) {
	c := NewItickWsClient("ws://example.test/crypto", "", "crypto", nil, nil, nil)

	first := cache.NormalizeClientMessage(types.ClientMessage{
		Topic:        types.TopicQuote,
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "BA",
	})
	second := cache.NormalizeClientMessage(types.ClientMessage{
		Topic:        types.TopicQuote,
		CategoryCode: "crypto",
		Symbol:       "ETHUSDT",
		Market:       "BA",
	})

	if err := c.ensureDesiredSubscriptions(map[string]types.ClientMessage{
		cache.BuildTopicKey(first): first,
	}); err == nil {
		t.Fatalf("expected sync to fail without a websocket connection")
	}
	if err := c.ensureDesiredSubscriptions(map[string]types.ClientMessage{
		cache.BuildTopicKey(second): second,
	}); err == nil {
		t.Fatalf("expected sync to fail without a websocket connection")
	}

	c.subMu.Lock()
	defer c.subMu.Unlock()

	if _, ok := c.desiredSubs[cache.BuildTopicKey(first)]; !ok {
		t.Fatalf("expected first subscription to remain desired")
	}
	if _, ok := c.desiredSubs[cache.BuildTopicKey(second)]; !ok {
		t.Fatalf("expected second subscription to be desired")
	}
}
