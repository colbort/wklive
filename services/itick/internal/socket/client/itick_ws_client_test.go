package client

import (
	"testing"

	"wklive/services/itick/internal/socket/server"
)

func TestEnsureDesiredSubscriptionsMergesWithoutReplacing(t *testing.T) {
	c := NewItickWsClient("ws://example.test/crypto", "", "crypto", nil, nil, nil, nil)

	first := server.NormalizeClientMessage(server.ClientMessage{
		Topic:        server.TopicQuote,
		CategoryCode: "crypto",
		Symbol:       "BTCUSDT",
		Market:       "BA",
	})
	second := server.NormalizeClientMessage(server.ClientMessage{
		Topic:        server.TopicQuote,
		CategoryCode: "crypto",
		Symbol:       "ETHUSDT",
		Market:       "BA",
	})

	if err := c.ensureDesiredSubscriptions(map[string]server.ClientMessage{
		server.BuildTopicKey(first): first,
	}); err == nil {
		t.Fatalf("expected sync to fail without a websocket connection")
	}
	if err := c.ensureDesiredSubscriptions(map[string]server.ClientMessage{
		server.BuildTopicKey(second): second,
	}); err == nil {
		t.Fatalf("expected sync to fail without a websocket connection")
	}

	c.subMu.Lock()
	defer c.subMu.Unlock()

	if _, ok := c.desiredSubs[server.BuildTopicKey(first)]; !ok {
		t.Fatalf("expected first subscription to remain desired")
	}
	if _, ok := c.desiredSubs[server.BuildTopicKey(second)]; !ok {
		t.Fatalf("expected second subscription to be desired")
	}
}
