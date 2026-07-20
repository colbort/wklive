package client

import (
	"encoding/json"
	"testing"

	"wklive/services/itick/internal/market/cache"
	"wklive/services/itick/internal/market/types"
)

func TestRawDecimalTokenPreservesSourcePrecision(t *testing.T) {
	got := rawDecimalToken(json.RawMessage(`{"ld":12345.678901234567890123}`), "ld")
	if got != "12345.678901234567890123" {
		t.Fatalf("raw decimal = %q", got)
	}
}

func TestReplaceDesiredSubscriptionsReplacesSnapshot(t *testing.T) {
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

	if err := c.replaceDesiredSubscriptions(map[string]types.ClientMessage{
		cache.BuildTopicKey(first):  first,
		cache.BuildTopicKey(second): second,
	}); err != nil {
		t.Fatalf("expected disconnected client to only store desired subscription: %v", err)
	}
	if err := c.replaceDesiredSubscriptions(map[string]types.ClientMessage{
		cache.BuildTopicKey(second): second,
	}); err != nil {
		t.Fatalf("expected disconnected client to only store desired subscription: %v", err)
	}

	c.subMu.Lock()
	defer c.subMu.Unlock()

	if _, ok := c.desiredSubs[cache.BuildTopicKey(first)]; ok {
		t.Fatalf("expected first subscription to be removed")
	}
	if _, ok := c.desiredSubs[cache.BuildTopicKey(second)]; !ok {
		t.Fatalf("expected second subscription to be desired")
	}
}

func TestBuildSubscriptionGroupsCombinesTypesWithSameProducts(t *testing.T) {
	c := NewItickWsClient("ws://example.test/crypto", "", "crypto", nil, nil, nil)
	items := make(map[string]types.ClientMessage)
	for _, symbol := range []string{"BTCUSDT", "ETHUSDT"} {
		for _, topic := range []types.Topic{types.TopicQuote, types.TopicTick} {
			msg := cache.NormalizeClientMessage(types.ClientMessage{
				Topic: topic, CategoryCode: "crypto", Symbol: symbol, Market: "BA",
			})
			items[cache.BuildTopicKey(msg)] = msg
		}
	}
	groups, err := c.buildSubscriptionGroups(items)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups["quote,tick"] != "BTCUSDT$BA,ETHUSDT$BA" {
		t.Fatalf("unexpected subscription groups: %#v", groups)
	}
}
