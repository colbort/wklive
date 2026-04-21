package client

import (
	"context"
	"encoding/json"

	"wklive/services/itick/internal/socket/server"

	"github.com/redis/go-redis/v9"
)

type ClusterEnvelope struct {
	Topic        string          `json:"topic"`
	CategoryCode string          `json:"categoryCode"`
	Symbol       string          `json:"symbol"`
	Market       string          `json:"market,omitempty"`
	Interval     string          `json:"interval,omitempty"`
	Payload      json.RawMessage `json:"payload"`
}

type ClusterBus struct {
	rdb     *redis.Client
	channel string
}

func NewClusterBus(rdb *redis.Client, channel string) *ClusterBus {
	return &ClusterBus{
		rdb:     rdb,
		channel: channel,
	}
}

func (b *ClusterBus) Publish(ctx context.Context, msg server.ClientMessage, payload any) error {
	msg = server.NormalizeClientMessage(msg)

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	env := ClusterEnvelope{
		Topic:        string(msg.Topic),
		CategoryCode: msg.CategoryCode,
		Symbol:       msg.Symbol,
		Market:       msg.Market,
		Interval:     msg.Interval,
		Payload:      raw,
	}

	bs, err := json.Marshal(env)
	if err != nil {
		return err
	}

	return b.rdb.Publish(ctx, b.channel, bs).Err()
}

func (b *ClusterBus) Subscribe(ctx context.Context, fn func(msg server.ClientMessage, payload any)) error {
	pubsub := b.rdb.Subscribe(ctx, b.channel)

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
			case item, ok := <-ch:
				if !ok {
					return
				}

				var env ClusterEnvelope
				if err := json.Unmarshal([]byte(item.Payload), &env); err != nil {
					continue
				}

				msg := server.ClientMessage{
					Topic:        server.Topic(env.Topic),
					CategoryCode: env.CategoryCode,
					Symbol:       env.Symbol,
					Market:       env.Market,
					Interval:     env.Interval,
				}
				msg = server.NormalizeClientMessage(msg)

				payload, err := decodeClusterPayload(msg.Topic, env.Payload)
				if err != nil {
					continue
				}

				fn(msg, payload)
			}
		}
	}()

	return nil
}

func decodeClusterPayload(topic server.Topic, raw json.RawMessage) (any, error) {
	switch topic {
	case server.TopicQuote:
		var v QuotePayload
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case server.TopicTick:
		var v TickPayload
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case server.TopicDepth:
		var v DepthPayload
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case server.TopicKline:
		var v KlinePayload
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil
	}

	return nil, nil
}
