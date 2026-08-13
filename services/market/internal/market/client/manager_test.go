package client

import (
	"context"
	"errors"
	"testing"

	"wklive/services/market/internal/market/provider"
)

type managerTestProvider struct {
	code       string
	categories []string
	streams    map[string]*managerTestStream
}

func (p *managerTestProvider) Code() string         { return p.code }
func (p *managerTestProvider) Categories() []string { return append([]string(nil), p.categories...) }
func (p *managerTestProvider) Supports(category string) bool {
	for _, item := range p.categories {
		if item == category {
			return true
		}
	}
	return false
}
func (*managerTestProvider) Warm(context.Context, []provider.Subscription) {}
func (*managerTestProvider) FetchQuote(context.Context, provider.Subscription) (*provider.Quote, error) {
	return nil, errors.New("not implemented")
}
func (p *managerTestProvider) NewStream(category string) (provider.Stream, error) {
	stream := &managerTestStream{}
	p.streams[category] = stream
	return stream, nil
}

type managerTestStream struct {
	reconnect      func(string)
	leader         bool
	replaced       int
	resubscribed   int
	resubscribeErr error
}

func (*managerTestStream) Start(context.Context)         {}
func (*managerTestStream) HasDesiredSubscriptions() bool { return true }
func (s *managerTestStream) ReplaceSubscriptions([]provider.Subscription) error {
	s.replaced++
	return nil
}
func (s *managerTestStream) Resubscribe(provider.Subscription) error {
	s.resubscribed++
	return s.resubscribeErr
}
func (s *managerTestStream) IsLeader() bool { return s.leader }
func (s *managerTestStream) SetReconnectHandler(handler func(string)) {
	s.reconnect = handler
}

func TestLoadCreatesOneStreamPerProviderAndCategory(t *testing.T) {
	itick := &managerTestProvider{code: "itick", categories: []string{"forex", "stock"}, streams: make(map[string]*managerTestStream)}
	traderMade := &managerTestProvider{code: "tradermade", categories: []string{"forex"}, streams: make(map[string]*managerTestStream)}
	manager := NewMarketManager(provider.NewComposite(itick, traderMade), nil, nil, nil, nil, StaleQuoteRecoveryConfig{})

	if err := manager.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(manager.clients) != 3 {
		t.Fatalf("stream count = %d, want 3", len(manager.clients))
	}
	for _, key := range []streamKey{
		{provider: "itick", category: "forex"},
		{provider: "itick", category: "stock"},
		{provider: "tradermade", category: "forex"},
	} {
		if manager.clients[key] == nil {
			t.Fatalf("missing stream %+v", key)
		}
	}
}

func TestReconnectIdentifiesProviderAndCategory(t *testing.T) {
	traderMade := &managerTestProvider{code: "tradermade", categories: []string{"forex"}, streams: make(map[string]*managerTestStream)}
	manager := NewMarketManager(provider.NewComposite(traderMade), nil, nil, nil, nil, StaleQuoteRecoveryConfig{})
	if err := manager.Load(context.Background()); err != nil {
		t.Fatal(err)
	}

	var gotProvider, gotCategory string
	manager.SetReconnectHandler(func(providerCode, category string) {
		gotProvider, gotCategory = providerCode, category
	})
	traderMade.streams["forex"].reconnect("forex")
	if gotProvider != "tradermade" || gotCategory != "forex" {
		t.Fatalf("reconnect identity = %s/%s", gotProvider, gotCategory)
	}
}

func TestResubscribeCategoryTargetsEveryLeaderAndAllowsPartialSupport(t *testing.T) {
	first := &managerTestStream{leader: true, resubscribeErr: errors.New("unsupported symbol")}
	second := &managerTestStream{leader: true}
	manager := &MarketManager{clients: map[streamKey]provider.Stream{
		{provider: "itick", category: "forex"}:      first,
		{provider: "tradermade", category: "forex"}: second,
	}}

	if err := manager.resubscribeCategory(provider.Subscription{CategoryCode: "forex", Symbol: "USDCNY"}); err != nil {
		t.Fatal(err)
	}
	if first.resubscribed != 1 || second.resubscribed != 1 {
		t.Fatalf("resubscribe counts = %d/%d", first.resubscribed, second.resubscribed)
	}
}
