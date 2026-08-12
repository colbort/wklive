package twelvedata

import (
	"context"
	"testing"

	"wklive/services/market/internal/market/provider"
	"wklive/services/market/internal/market/types"
)

func TestReplaceSubscriptionsDeduplicatesInternalTopics(t *testing.T) {
	stream := newStream("wss://example.invalid", "key", nil, testSymbolCatalog(), nil)
	base := provider.Subscription{CategoryCode: "forex", Market: "GB", Symbol: "usd/cny"}
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
	stream := newStream("wss://example.invalid", "key", nil, testSymbolCatalog(), nil)
	stream.sent["USDCNY"] = struct{}{}
	stream.handleSubscriptionACK(wsEnvelope{
		Event: "subscribe-status",
		Fails: []wsStatusItem{{Symbol: "USD/CNY", Code: 400, Message: "symbol unavailable"}},
	})
	if _, ok := stream.sent["USDCNY"]; ok {
		t.Fatal("rejected subscription remained marked as sent")
	}
}

func TestPriceEventUsesSourceTimestamp(t *testing.T) {
	stream := newStream("wss://example.invalid", "key", nil, testSymbolCatalog(), nil)
	stream.desired["USDCNY"] = provider.Subscription{CategoryCode: "forex", Market: "GB", Symbol: "USDCNY"}
	// A nil cache would panic only if a known, valid message reaches publishing;
	// malformed timestamps must be rejected before that point.
	stream.publishPrice(context.Background(), wsEnvelope{Event: "price", Symbol: "USD/CNY", Timestamp: 0, Price: []byte(`7.18`)})
}

func TestReplaceSubscriptionsSkipsSymbolsMissingFromCatalog(t *testing.T) {
	stream := newStream("wss://example.invalid", "key", nil, testSymbolCatalog(), nil)
	if err := stream.ReplaceSubscriptions([]provider.Subscription{
		{Topic: types.TopicQuote, CategoryCode: "forex", Symbol: "USDCNY"},
		{Topic: types.TopicQuote, CategoryCode: "forex", Symbol: "COFFEE"},
	}); err != nil {
		t.Fatal(err)
	}
	if len(stream.desired) != 1 {
		t.Fatalf("desired subscriptions = %d, want only catalog-supported USDCNY", len(stream.desired))
	}
}

func TestStreamRemainsStartableWhenCatalogLoadInitiallyFailed(t *testing.T) {
	catalog := &SymbolCatalog{symbols: make(map[string]string)}
	stream := newStream("wss://example.invalid", "key", nil, catalog, nil)
	if err := stream.ReplaceSubscriptions([]provider.Subscription{
		{Topic: types.TopicQuote, CategoryCode: "forex", Symbol: "USDCNY"},
	}); err != nil {
		t.Fatal(err)
	}
	if !stream.HasDesiredSubscriptions() {
		t.Fatal("stream would not start and retry an initially failed catalog load")
	}
}

func testSymbolCatalog() *SymbolCatalog {
	return &SymbolCatalog{loaded: true, symbols: map[string]string{"USDCNY": "USD/CNY"}}
}
