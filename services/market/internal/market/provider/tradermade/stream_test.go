package tradermade

import (
	"testing"

	"wklive/services/market/internal/market/provider"
	"wklive/services/market/internal/market/types"
)

func TestReplaceSubscriptionsDeduplicatesInternalTopics(t *testing.T) {
	stream := newStream("wss://example.invalid", "key", false, nil, loadedStreamCatalog(map[string]string{"USDCNY": "USDCNY"}), nil)
	base := provider.Subscription{CategoryCode: "forex", Market: "gb", Symbol: "usd/cny"}
	items := make([]provider.Subscription, 0, 4)
	for _, topic := range []types.Topic{types.TopicQuote, types.TopicDepth, types.TopicTick, types.TopicKline} {
		item := base
		item.Topic = topic
		items = append(items, item)
	}
	if err := stream.ReplaceSubscriptions(items); err != nil {
		t.Fatal(err)
	}
	if len(stream.desired) != 1 {
		t.Fatalf("desired subscriptions = %d, want 1", len(stream.desired))
	}
	if got := stream.desired["USDCNY"].Topic; got != types.TopicQuote {
		t.Fatalf("preferred internal topic = %q", got)
	}
}

func TestSubscriptionACKRemovesRejectedSymbolFromSent(t *testing.T) {
	stream := newStream("wss://example.invalid", "key", false, nil, loadedStreamCatalog(map[string]string{"USDCNY": "USDCNY"}), nil)
	stream.sent["USDCNY"] = struct{}{}
	stream.handleSubscriptionACK(wsEnvelope{
		Type:          "sub_ack",
		Denied:        []string{"USDCNY:QUOTE"},
		DeniedReasons: map[string]string{"USDCNY:QUOTE": "symbol_limit"},
	})
	if _, ok := stream.sent["USDCNY"]; ok {
		t.Fatal("rejected subscription remained marked as sent")
	}
}

func TestCFDAliasUsesUpstreamKeyAndPreservesInternalSymbol(t *testing.T) {
	stream := newStream("wss://example.invalid", "key", false, nil, loadedStreamCatalog(map[string]string{
		"AAPLUSD": "AAPLUSD",
		"AAPL":    "AAPLUSD",
	}), nil)
	if err := stream.ReplaceSubscriptions([]provider.Subscription{{CategoryCode: "forex", Symbol: "AAPL"}}); err != nil {
		t.Fatal(err)
	}
	item, ok := stream.desired["AAPLUSD"]
	if !ok {
		t.Fatal("CFD alias did not use exact upstream symbol as desired key")
	}
	if item.Symbol != "AAPL" {
		t.Fatalf("internal cache symbol = %q, want AAPL", item.Symbol)
	}
}

func TestRequestedSubscriptionsSurviveTransientCatalogFailure(t *testing.T) {
	stream := newStream("wss://example.invalid", "key", false, nil, newStreamSymbolCatalog("https://example.invalid", "key", nil), nil)
	if err := stream.ReplaceSubscriptions([]provider.Subscription{{CategoryCode: "forex", Symbol: "USDCNY"}}); err != nil {
		t.Fatal(err)
	}
	if !stream.HasDesiredSubscriptions() {
		t.Fatal("stream should remain startable while its symbol catalog is unavailable")
	}
	if len(stream.desired) != 0 || len(stream.requested) != 1 {
		t.Fatalf("desired=%d requested=%d", len(stream.desired), len(stream.requested))
	}
}

func TestWebsocketDepthUsesLadderWhenAvailable(t *testing.T) {
	payload := websocketDepthPayload(wsEnvelope{
		Ask: "7.20", Bid: "7.10",
		Ladder: &wsLadder{
			Asks: [][]string{{"7.21", "200"}, {"7.20", "100"}},
			Bids: [][]string{{"7.09", "400"}, {"7.10", "300"}},
		},
	})
	if len(payload.Asks) != 2 || payload.Asks[0].Price != 7.20 || payload.Asks[0].Volume != 100 {
		t.Fatalf("unexpected asks: %+v", payload.Asks)
	}
	if len(payload.Bids) != 2 || payload.Bids[0].Price != 7.10 || payload.Bids[0].Volume != 300 {
		t.Fatalf("unexpected bids: %+v", payload.Bids)
	}
}

func loadedStreamCatalog(symbols map[string]string) *SymbolCatalog {
	catalog := newStreamSymbolCatalog("https://example.invalid", "key", nil)
	for internal, upstream := range symbols {
		addSymbolMapping(catalog.forward, catalog.reverse, internal, upstream)
	}
	catalog.loaded = true
	return catalog
}
