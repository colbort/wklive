package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Composite runs every configured provider that supports a category. Providers
// keep independent connections and authorities while sharing the normalized
// subscription lifecycle expected by MarketManager.
type Composite struct {
	providers []RealtimeProvider
}

func NewComposite(providers ...RealtimeProvider) *Composite {
	items := make([]RealtimeProvider, 0, len(providers))
	seen := make(map[string]struct{})
	for _, item := range providers {
		if item == nil {
			continue
		}
		code := strings.ToLower(strings.TrimSpace(item.Code()))
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		items = append(items, item)
	}
	return &Composite{providers: items}
}

func (c *Composite) Code() string { return "multi" }

func (c *Composite) Supports(category string) bool {
	return len(c.supporting(category)) > 0
}

func (c *Composite) Warm(ctx context.Context, subscriptions []Subscription) {
	var wait sync.WaitGroup
	for _, item := range c.providers {
		items := subscriptionsForProvider(item, subscriptions)
		if len(items) == 0 {
			continue
		}
		wait.Add(1)
		go func(selected RealtimeProvider, selectedItems []Subscription) {
			defer wait.Done()
			selected.Warm(ctx, selectedItems)
		}(item, items)
	}
	wait.Wait()
}

func (c *Composite) FetchQuote(ctx context.Context, subscription Subscription) (*Quote, error) {
	items := c.supporting(subscription.CategoryCode)
	if len(items) == 0 {
		return nil, fmt.Errorf("no realtime provider supports category %q", subscription.CategoryCode)
	}
	type result struct {
		quote *Quote
		err   error
	}
	results := make(chan result, len(items))
	for _, item := range items {
		go func(selected RealtimeProvider) {
			quote, err := selected.FetchQuote(ctx, subscription)
			if err != nil {
				err = fmt.Errorf("%s: %w", selected.Code(), err)
			}
			results <- result{quote: quote, err: err}
		}(item)
	}
	var freshest *Quote
	providerErrors := make([]error, 0, len(items))
	for range items {
		result := <-results
		if result.err != nil {
			providerErrors = append(providerErrors, result.err)
			continue
		}
		if result.quote != nil && (freshest == nil || result.quote.Ts > freshest.Ts) {
			freshest = result.quote
		}
	}
	if freshest != nil {
		return freshest, nil
	}
	return nil, fmt.Errorf("all realtime providers failed for category %q: %w", subscription.CategoryCode, errors.Join(providerErrors...))
}

func (c *Composite) NewStream(category string) (Stream, error) {
	items := c.supporting(category)
	streams := make([]Stream, 0, len(items))
	var providerErrors []error
	for _, item := range items {
		stream, err := item.NewStream(category)
		if err != nil {
			providerErrors = append(providerErrors, fmt.Errorf("%s: %w", item.Code(), err))
			continue
		}
		streams = append(streams, stream)
	}
	if len(streams) == 0 {
		return nil, fmt.Errorf("no realtime stream available for category %q: %w", category, errors.Join(providerErrors...))
	}
	return &compositeStream{streams: streams}, nil
}

func (c *Composite) supporting(category string) []RealtimeProvider {
	if c == nil {
		return nil
	}
	items := make([]RealtimeProvider, 0, len(c.providers))
	for _, item := range c.providers {
		if item.Supports(category) {
			items = append(items, item)
		}
	}
	return items
}

func subscriptionsForProvider(item RealtimeProvider, subscriptions []Subscription) []Subscription {
	items := make([]Subscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		if item.Supports(subscription.CategoryCode) {
			items = append(items, subscription)
		}
	}
	return items
}

type compositeStream struct {
	streams []Stream
}

func (s *compositeStream) Start(ctx context.Context) {
	for _, stream := range s.streams {
		stream.Start(ctx)
	}
}

func (s *compositeStream) HasDesiredSubscriptions() bool {
	for _, stream := range s.streams {
		if stream.HasDesiredSubscriptions() {
			return true
		}
	}
	return false
}

func (s *compositeStream) ReplaceSubscriptions(items []Subscription) error {
	providerErrors := make([]error, 0)
	for _, stream := range s.streams {
		if err := stream.ReplaceSubscriptions(items); err != nil {
			providerErrors = append(providerErrors, err)
		}
	}
	return errors.Join(providerErrors...)
}

func (s *compositeStream) Resubscribe(item Subscription) error {
	providerErrors := make([]error, 0)
	leaders := 0
	for _, stream := range s.streams {
		if !stream.IsLeader() {
			continue
		}
		leaders++
		if err := stream.Resubscribe(item); err != nil {
			providerErrors = append(providerErrors, err)
		}
	}
	if leaders == 0 {
		return errors.New("no provider stream is leader")
	}
	return errors.Join(providerErrors...)
}

func (s *compositeStream) IsLeader() bool {
	for _, stream := range s.streams {
		if stream.IsLeader() {
			return true
		}
	}
	return false
}

func (s *compositeStream) SetReconnectHandler(handler func(category string)) {
	for _, stream := range s.streams {
		stream.SetReconnectHandler(handler)
	}
}

var _ RealtimeProvider = (*Composite)(nil)
var _ Stream = (*compositeStream)(nil)
