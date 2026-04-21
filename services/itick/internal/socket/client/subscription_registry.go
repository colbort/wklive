package client

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"wklive/services/itick/internal/socket/server"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
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
	return server.NormalizeClientMessage(server.ClientMessage{
		Topic:        server.Topic(c.Topic),
		CategoryCode: c.CategoryCode,
		Symbol:       c.Symbol,
		Market:       c.Market,
		Interval:     c.Interval,
	})
}

type SubscriptionRegistry struct {
	rdb           *redis.Client
	hashKeyPrefix string
	changeChannel string
	ownerID       string
	leaseTTL      time.Duration

	mu        sync.Mutex
	refCounts map[string]int
	renewers  map[string]context.CancelFunc
}

func NewSubscriptionRegistry(rdb *redis.Client, hashKeyPrefix, changeChannel string) *SubscriptionRegistry {
	return &SubscriptionRegistry{
		rdb:           rdb,
		hashKeyPrefix: strings.TrimRight(hashKeyPrefix, ":"),
		changeChannel: changeChannel,
		ownerID:       buildRegistryOwnerID(),
		leaseTTL:      90 * time.Second,
		refCounts:     make(map[string]int),
		renewers:      make(map[string]context.CancelFunc),
	}
}

func (r *SubscriptionRegistry) categoryIndexKey(categoryCode string) string {
	return fmt.Sprintf("%s:index:%s", r.hashKeyPrefix, normalizeLeasePart(categoryCode))
}

func (r *SubscriptionRegistry) topicLeaseSetKey(categoryCode, topicKey string) string {
	return fmt.Sprintf("%s:lease:%s:%s", r.hashKeyPrefix, normalizeLeasePart(categoryCode), shortHash(topicKey))
}

func (r *SubscriptionRegistry) ownerLeaseKey(categoryCode, topicKey string) string {
	return fmt.Sprintf("%s:lease:%s:%s:%s", r.hashKeyPrefix, normalizeLeasePart(categoryCode), shortHash(topicKey), r.ownerID)
}

func (r *SubscriptionRegistry) Add(ctx context.Context, msg server.ClientMessage) error {
	msg = server.NormalizeClientMessage(msg)
	topicKey := server.BuildTopicKey(msg)
	indexKey := r.categoryIndexKey(msg.CategoryCode)
	setKey := r.topicLeaseSetKey(msg.CategoryCode, topicKey)
	leaseKey := r.ownerLeaseKey(msg.CategoryCode, topicKey)

	if !r.acquireLocalRef(topicKey) {
		return nil
	}

	before, err := r.activeLeaseCount(ctx, setKey)
	if err != nil {
		r.releaseLocalRef(topicKey)
		return err
	}

	if err := r.writeLease(ctx, indexKey, setKey, leaseKey, topicKey); err != nil {
		r.releaseLocalRef(topicKey)
		return err
	}

	r.startLeaseRenewer(indexKey, setKey, leaseKey, topicKey)

	after, err := r.activeLeaseCount(ctx, setKey)
	if err != nil {
		return err
	}

	if before == 0 && after > 0 {
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
	msg = server.NormalizeClientMessage(msg)
	topicKey := server.BuildTopicKey(msg)
	indexKey := r.categoryIndexKey(msg.CategoryCode)
	setKey := r.topicLeaseSetKey(msg.CategoryCode, topicKey)
	leaseKey := r.ownerLeaseKey(msg.CategoryCode, topicKey)

	if !r.releaseLocalRef(topicKey) {
		return nil
	}

	r.stopLeaseRenewer(leaseKey)

	before, err := r.activeLeaseCount(ctx, setKey)
	if err != nil {
		return err
	}

	if _, err := r.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, leaseKey)
		pipe.SRem(ctx, setKey, leaseKey)
		return nil
	}); err != nil {
		return err
	}

	after, err := r.activeLeaseCount(ctx, setKey)
	if err != nil {
		return err
	}

	if after == 0 {
		_ = r.rdb.SRem(ctx, indexKey, topicKey).Err()
	}

	if before > 0 && after == 0 {
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

func (r *SubscriptionRegistry) EnsureLease(ctx context.Context, msg server.ClientMessage) error {
	msg = server.NormalizeClientMessage(msg)
	topicKey := server.BuildTopicKey(msg)
	indexKey := r.categoryIndexKey(msg.CategoryCode)
	setKey := r.topicLeaseSetKey(msg.CategoryCode, topicKey)
	leaseKey := r.ownerLeaseKey(msg.CategoryCode, topicKey)

	if err := r.writeLease(ctx, indexKey, setKey, leaseKey, topicKey); err != nil {
		return err
	}

	r.startLeaseRenewer(indexKey, setKey, leaseKey, topicKey)
	return nil
}

func (r *SubscriptionRegistry) EnsureAndNotify(ctx context.Context, msg server.ClientMessage) error {
	msg = server.NormalizeClientMessage(msg)
	if err := r.EnsureLease(ctx, msg); err != nil {
		return err
	}

	return r.publishChange(ctx, SubscriptionChange{
		Action:       SubscriptionAdd,
		Topic:        string(msg.Topic),
		CategoryCode: msg.CategoryCode,
		Symbol:       msg.Symbol,
		Market:       msg.Market,
		Interval:     msg.Interval,
	})
}

func (r *SubscriptionRegistry) acquireLocalRef(topicKey string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refCounts[topicKey]++
	return r.refCounts[topicKey] == 1
}

func (r *SubscriptionRegistry) releaseLocalRef(topicKey string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := r.refCounts[topicKey]
	if count <= 0 {
		return false
	}

	if count > 1 {
		r.refCounts[topicKey] = count - 1
		return false
	}

	delete(r.refCounts, topicKey)
	return true
}

func (r *SubscriptionRegistry) ListActive(ctx context.Context, categoryCode string) ([]server.ClientMessage, error) {
	indexKey := r.categoryIndexKey(categoryCode)

	topicKeys, err := r.rdb.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}

	out := make([]server.ClientMessage, 0, len(topicKeys))
	for _, topicKey := range topicKeys {
		setKey := r.topicLeaseSetKey(categoryCode, topicKey)
		cnt, err := r.activeLeaseCount(ctx, setKey)
		if err != nil {
			return nil, err
		}
		if cnt == 0 {
			_ = r.rdb.SRem(ctx, indexKey, topicKey).Err()
			continue
		}
		out = append(out, server.ParseTopicKey(topicKey))
	}

	return out, nil
}

func (r *SubscriptionRegistry) writeLease(ctx context.Context, indexKey, setKey, leaseKey, topicKey string) error {
	_, err := r.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, leaseKey, topicKey, r.leaseTTL)
		pipe.SAdd(ctx, setKey, leaseKey)
		pipe.Expire(ctx, setKey, r.leaseTTL*2)
		pipe.SAdd(ctx, indexKey, topicKey)
		pipe.Expire(ctx, indexKey, r.leaseTTL*2)
		return nil
	})
	return err
}

func (r *SubscriptionRegistry) activeLeaseCount(ctx context.Context, setKey string) (int64, error) {
	leaseKeys, err := r.rdb.SMembers(ctx, setKey).Result()
	if err != nil {
		return 0, err
	}
	if len(leaseKeys) == 0 {
		return 0, nil
	}

	var active int64
	stale := make([]interface{}, 0)
	for _, leaseKey := range leaseKeys {
		exists, err := r.rdb.Exists(ctx, leaseKey).Result()
		if err != nil {
			return 0, err
		}
		if exists > 0 {
			active++
			continue
		}
		stale = append(stale, leaseKey)
	}

	if len(stale) > 0 {
		_ = r.rdb.SRem(ctx, setKey, stale...).Err()
	}

	return active, nil
}

func (r *SubscriptionRegistry) startLeaseRenewer(indexKey, setKey, leaseKey, topicKey string) {
	r.mu.Lock()
	if cancel, ok := r.renewers[leaseKey]; ok {
		cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.renewers[leaseKey] = cancel
	r.mu.Unlock()

	interval := r.leaseTTL / 3
	if interval <= 0 {
		interval = 30 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.writeLease(ctx, indexKey, setKey, leaseKey, topicKey); err != nil {
					logx.Errorf("refresh subscription lease failed: lease=%s err=%v", leaseKey, err)
				}
			}
		}
	}()
}

func (r *SubscriptionRegistry) stopLeaseRenewer(leaseKey string) {
	r.mu.Lock()
	cancel := r.renewers[leaseKey]
	delete(r.renewers, leaseKey)
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
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

func buildRegistryOwnerID() string {
	hostname, _ := os.Hostname()
	raw := fmt.Sprintf("%s:%d:%d", hostname, os.Getpid(), time.Now().UnixNano())
	return shortHash(raw)
}

func normalizeLeasePart(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func shortHash(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}
