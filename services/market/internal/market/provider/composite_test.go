package provider

import (
	"context"
	"errors"
	"testing"
)

type compositeTestProvider struct {
	code       string
	categories map[string]bool
	quote      *Quote
	warmed     []Subscription
	stream     *compositeTestStream
}

func (p *compositeTestProvider) Code() string { return p.code }
func (p *compositeTestProvider) Categories() []string {
	items := make([]string, 0, len(p.categories))
	for category := range p.categories {
		items = append(items, category)
	}
	return items
}
func (p *compositeTestProvider) Supports(category string) bool {
	return p.categories[category]
}
func (p *compositeTestProvider) Warm(_ context.Context, items []Subscription) {
	p.warmed = append(p.warmed, items...)
}
func (p *compositeTestProvider) FetchQuote(context.Context, Subscription) (*Quote, error) {
	if p.quote == nil {
		return nil, errors.New("no quote")
	}
	return p.quote, nil
}
func (p *compositeTestProvider) NewStream(string) (Stream, error) {
	if p.stream == nil {
		return nil, errors.New("no stream")
	}
	return p.stream, nil
}

type compositeTestStream struct {
	replaced int
	leader   bool
}

func (*compositeTestStream) Start(context.Context)                       {}
func (*compositeTestStream) HasDesiredSubscriptions() bool               { return true }
func (s *compositeTestStream) ReplaceSubscriptions([]Subscription) error { s.replaced++; return nil }
func (*compositeTestStream) Resubscribe(Subscription) error              { return nil }
func (s *compositeTestStream) IsLeader() bool                            { return s.leader }
func (*compositeTestStream) SetReconnectHandler(func(string))            {}

func TestCompositeRunsAllSupportingProviders(t *testing.T) {
	firstStream := &compositeTestStream{leader: true}
	secondStream := &compositeTestStream{leader: true}
	first := &compositeTestProvider{code: "itick", categories: map[string]bool{"forex": true}, stream: firstStream}
	second := &compositeTestProvider{code: "tradermade", categories: map[string]bool{"forex": true}, stream: secondStream}
	composite := NewComposite(first, second)

	stream, err := composite.NewStream("forex")
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.ReplaceSubscriptions([]Subscription{{CategoryCode: "forex", Symbol: "USDCNY"}}); err != nil {
		t.Fatal(err)
	}
	if firstStream.replaced != 1 || secondStream.replaced != 1 {
		t.Fatalf("subscriptions were not broadcast: first=%d second=%d", firstStream.replaced, secondStream.replaced)
	}
}

func TestCompositeFetchQuoteChoosesFreshestProvider(t *testing.T) {
	first := &compositeTestProvider{code: "itick", categories: map[string]bool{"forex": true}, quote: &Quote{Ts: 100}}
	second := &compositeTestProvider{code: "tradermade", categories: map[string]bool{"forex": true}, quote: &Quote{Ts: 200}}
	quote, err := NewComposite(first, second).FetchQuote(context.Background(), Subscription{CategoryCode: "forex"})
	if err != nil {
		t.Fatal(err)
	}
	if quote.Ts != 200 {
		t.Fatalf("selected timestamp = %d", quote.Ts)
	}
}

func TestCompositeExposesIndependentProvidersAndCategories(t *testing.T) {
	first := &compositeTestProvider{code: "itick", categories: map[string]bool{"stock": true, "forex": true}}
	second := &compositeTestProvider{code: "tradermade", categories: map[string]bool{"forex": true}}
	composite := NewComposite(first, second)
	if got := composite.Categories(); len(got) != 2 || got[0] != "forex" || got[1] != "stock" {
		t.Fatalf("categories = %v", got)
	}
	if got := Sources(composite); len(got) != 2 || got[0].Code() != "itick" || got[1].Code() != "tradermade" {
		t.Fatalf("sources = %v", got)
	}
}
