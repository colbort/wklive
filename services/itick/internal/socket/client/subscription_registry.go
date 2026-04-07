package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"wklive/services/itick/internal/socket/server"

	"github.com/redis/go-redis/v9"
)

type SubscriptionChangeAction string

const (
	SubscriptionAdd    SubscriptionChangeAction = "add"
	SubscriptionRemove SubscriptionChangeAction = "remove"
)

type SubscriptionChange struct {
	Action       SubscriptionChangeAction `json:"action"`
	Topic        string                   `json:"topic"`
	CategoryCode string                   `json:"categoryCode"`
	Symbol       string                   `json:"symbol"`
	Market       string                   `json:"market"`
	Interval     string                   `json:"interval,omitempty"`
}

func (c SubscriptionChange) ToClientMessage() server.ClientMessage {
	return server.ClientMessage{
		Topic:        server.Topic(c.Topic),
		CategoryCode: c.CategoryCode,
		Symbol:       c.Symbol,
		Market:       c.Market,
		Interval:     c.Interval,
	}
}

type SubscriptionRegistry struct {
	rdb           *redis.Client
	hashKeyPrefix string
	changeChannel string
}

func NewSubscriptionRegistry(rdb *redis.Client, hashKeyPrefix, changeChannel string) *SubscriptionRegistry {
	return &SubscriptionRegistry{
		rdb:           rdb,
		hashKeyPrefix: strings.TrimRight(hashKeyPrefix, ":"),
		changeChannel: changeChannel,
	}
}

func (r *SubscriptionRegistry) hashKey(categoryCode string) string {
	return fmt.Sprintf("%s:%s", r.hashKeyPrefix, strings.ToLower(strings.TrimSpace(categoryCode)))
}

func (r *SubscriptionRegistry) Add(ctx context.Context, msg server.ClientMessage) error {
	key := r.hashKey(msg.CategoryCode)
	field := server.BuildTopicKey(msg)

	n, err := r.rdb.HIncrBy(ctx, key, field, 1).Result()
	if err != nil {
		return err
	}

	if n == 1 {
		change := SubscriptionChange{
			Action:       SubscriptionAdd,
			Topic:        string(msg.Topic),
			CategoryCode: msg.CategoryCode,
			Symbol:       msg.Symbol,
			Market:       msg.Market,
			Interval:     msg.Interval,
		}
		return r.publishChange(ctx, change)
	}

	return nil
}

func (r *SubscriptionRegistry) Remove(ctx context.Context, msg server.ClientMessage) error {
	key := r.hashKey(msg.CategoryCode)
	field := server.BuildTopicKey(msg)

	n, err := r.rdb.HIncrBy(ctx, key, field, -1).Result()
	if err != nil {
		return err
	}

	if n <= 0 {
		if err := r.rdb.HDel(ctx, key, field).Err(); err != nil {
			return err
		}

		change := SubscriptionChange{
			Action:       SubscriptionRemove,
			Topic:        string(msg.Topic),
			CategoryCode: msg.CategoryCode,
			Symbol:       msg.Symbol,
			Market:       msg.Market,
			Interval:     msg.Interval,
		}
		return r.publishChange(ctx, change)
	}

	return nil
}

func (r *SubscriptionRegistry) ListActive(ctx context.Context, categoryCode string) ([]server.ClientMessage, error) {
	key := r.hashKey(categoryCode)

	m, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	out := make([]server.ClientMessage, 0, len(m))
	for field, val := range m {
		cnt, _ := strconv.ParseInt(val, 10, 64)
		if cnt <= 0 {
			continue
		}
		out = append(out, server.ParseTopicKey(field))
	}

	return out, nil
}

func (r *SubscriptionRegistry) WatchChanges(ctx context.Context, fn func(change SubscriptionChange)) error {
	pubsub := r.rdb.Subscribe(ctx, r.changeChannel)

	_, err := pubsub.Receive(ctx)
	if err != nil {
		_ = pubsub.Close()
		return err
	}

	go func() {
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var change SubscriptionChange
				if err := json.Unmarshal([]byte(msg.Payload), &change); err != nil {
					continue
				}

				fn(change)
			}
		}
	}()

	return nil
}

func (r *SubscriptionRegistry) publishChange(ctx context.Context, change SubscriptionChange) error {
	bs, err := json.Marshal(change)
	if err != nil {
		return err
	}
	return r.rdb.Publish(ctx, r.changeChannel, bs).Err()
}
