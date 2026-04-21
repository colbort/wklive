package client

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"wklive/services/itick/internal/socket/server"
	"wklive/services/itick/models"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type ItickManager struct {
	wsUrl string
	token string

	model models.ItickCategoryModel
	hub   *server.Hub

	busRedis  *redis.Client
	lockRedis *redis.Client
	registry  *SubscriptionRegistry
	bus       *ClusterBus

	mu      sync.RWMutex
	clients map[string]*ItickWsClient
}

func NewItickManager(
	wsUrl string,
	token string,
	hub *server.Hub,
	model models.ItickCategoryModel,
	busRedis *redis.Client,
	lockRedis *redis.Client,
) *ItickManager {
	registry := NewSubscriptionRegistry(
		busRedis,
		"itick:subs",
		"itick:subs:changes",
	)

	bus := NewClusterBus(
		busRedis,
		"itick:cluster:bus",
	)

	return &ItickManager{
		wsUrl:     wsUrl,
		token:     token,
		model:     model,
		hub:       hub,
		busRedis:  busRedis,
		lockRedis: lockRedis,
		registry:  registry,
		bus:       bus,
		clients:   make(map[string]*ItickWsClient),
	}
}

func (m *ItickManager) Load(ctx context.Context) error {
	categories, err := m.model.FindAll(ctx)
	if err != nil {
		return err
	}

	newClients := make(map[string]*ItickWsClient)

	for _, item := range categories {
		categoryCode := strings.ToLower(strings.TrimSpace(item.CategoryCode))
		wsURL := strings.TrimSpace(m.wsUrl)

		if categoryCode == "" || wsURL == "" {
			logx.Errorf("skip invalid itick category, code=%s, wsURL=%s", item.CategoryCode, m.wsUrl)
			continue
		}

		upstreamURL := fmt.Sprintf("%s/%s", wsURL, categoryCode)
		lockKey := "itick:leader:" + sha1Hex(upstreamURL)

		newClients[categoryCode] = NewItickWsClient(
			upstreamURL,
			m.token,
			categoryCode,
			m.bus,
			m.registry,
			NewRedisLeaderLock(m.lockRedis, lockKey),
			m.hub,
		)
	}

	if len(newClients) == 0 {
		return fmt.Errorf("no available itick categories found")
	}

	m.mu.Lock()
	m.clients = newClients
	m.mu.Unlock()

	logx.Infof("itick manager loaded categories success, count=%d", len(newClients))
	return nil
}

func (m *ItickManager) Start(ctx context.Context) error {
	m.hub.SetHooks(
		func(key string, msg server.ClientMessage) {
			hookCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := m.EnsureUpstreamSubscription(hookCtx, msg); err != nil {
				logx.Errorf("ensure upstream subscription failed, topic=%s err=%v", key, err)
			}
		},
		nil,
	)

	if err := m.bus.Subscribe(ctx, func(msg server.ClientMessage, payload any) {
		m.hub.Broadcast(msg, payload)
	}); err != nil {
		return err
	}

	if err := m.registry.WatchChanges(ctx, func(change SubscriptionChange) {
		m.mu.RLock()
		cli := m.clients[strings.ToLower(strings.TrimSpace(change.CategoryCode))]
		m.mu.RUnlock()

		if cli != nil {
			cli.HandleSubscriptionChange(change)
		}
	}); err != nil {
		return err
	}

	m.mu.RLock()
	clients := make([]*ItickWsClient, 0, len(m.clients))
	for _, cli := range m.clients {
		clients = append(clients, cli)
	}
	m.mu.RUnlock()

	for _, cli := range clients {
		cli.Start(ctx)
	}

	return nil
}

func (m *ItickManager) AddGlobalSubscription(ctx context.Context, msg server.ClientMessage) error {
	return m.registry.Add(ctx, msg)
}

func (m *ItickManager) RemoveGlobalSubscription(ctx context.Context, msg server.ClientMessage) error {
	return m.registry.Remove(ctx, msg)
}

func (m *ItickManager) EnsureUpstreamSubscription(ctx context.Context, msg server.ClientMessage) error {
	if err := m.registry.EnsureAndNotify(ctx, msg); err != nil {
		return err
	}

	m.mu.RLock()
	cli := m.clients[strings.ToLower(strings.TrimSpace(msg.CategoryCode))]
	m.mu.RUnlock()

	if cli != nil && cli.IsLeader() {
		return cli.subscribeByClientMessage(msg)
	}

	return nil
}

func sha1Hex(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
