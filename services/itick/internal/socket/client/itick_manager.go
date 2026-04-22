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

	startMu sync.Mutex
	started bool

	changeMu            sync.Mutex
	pendingChanges      []SubscriptionChange
	pendingChangesTimer *time.Timer
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
	m.startMu.Lock()
	if m.started {
		m.startMu.Unlock()
		return nil
	}
	m.started = true
	m.startMu.Unlock()

	if err := m.bus.Subscribe(ctx, func(msg server.ClientMessage, payload any) {
		m.hub.Broadcast(msg, payload)
	}); err != nil {
		m.startMu.Lock()
		m.started = false
		m.startMu.Unlock()
		return err
	}

	if err := m.registry.WatchChanges(ctx, func(changes []SubscriptionChange) {
		m.queueSubscriptionChanges(changes)
	}); err != nil {
		m.startMu.Lock()
		m.started = false
		m.startMu.Unlock()
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

func (m *ItickManager) queueSubscriptionChanges(changes []SubscriptionChange) {
	if len(changes) == 0 {
		return
	}

	m.changeMu.Lock()
	defer m.changeMu.Unlock()

	m.pendingChanges = append(m.pendingChanges, changes...)
	if m.pendingChangesTimer != nil {
		m.pendingChangesTimer.Reset(defaultSubscribeDelay)
		return
	}

	m.pendingChangesTimer = time.AfterFunc(defaultSubscribeDelay, m.flushSubscriptionChanges)
}

func (m *ItickManager) flushSubscriptionChanges() {
	m.changeMu.Lock()
	changes := m.pendingChanges
	m.pendingChanges = nil
	m.pendingChangesTimer = nil
	m.changeMu.Unlock()

	if len(changes) == 0 {
		return
	}

	byCategory := make(map[string][]SubscriptionChange)
	for _, change := range changes {
		categoryCode := strings.ToLower(strings.TrimSpace(change.CategoryCode))
		if categoryCode == "" {
			continue
		}
		byCategory[categoryCode] = append(byCategory[categoryCode], change)
	}

	m.mu.RLock()
	clients := make(map[string]*ItickWsClient, len(byCategory))
	for categoryCode := range byCategory {
		clients[categoryCode] = m.clients[categoryCode]
	}
	m.mu.RUnlock()

	for categoryCode, changes := range byCategory {
		cli := clients[categoryCode]
		if cli != nil {
			cli.HandleSubscriptionChanges(changes)
		}
	}
}

func (m *ItickManager) AddGlobalSubscriptions(ctx context.Context, msgs []server.ClientMessage) error {
	return m.registry.AddMany(ctx, msgs)
}

func (m *ItickManager) RemoveGlobalSubscriptions(ctx context.Context, msgs []server.ClientMessage) error {
	return m.registry.RemoveMany(ctx, msgs)
}

func (m *ItickManager) EnsureUpstreamSubscriptions(ctx context.Context, msgs []server.ClientMessage) error {
	uniqueMsgs := normalizeUniqueMessages(msgs)
	if len(uniqueMsgs) == 0 {
		return nil
	}

	byCategory := make(map[string]map[string]server.ClientMessage)
	if err := m.registry.EnsureLeases(ctx, uniqueMsgs); err != nil {
		return err
	}
	for _, msg := range uniqueMsgs {
		categoryCode := strings.ToLower(strings.TrimSpace(msg.CategoryCode))
		if byCategory[categoryCode] == nil {
			byCategory[categoryCode] = make(map[string]server.ClientMessage)
		}
		byCategory[categoryCode][server.BuildTopicKey(msg)] = msg
	}

	m.mu.RLock()
	clients := make(map[string]*ItickWsClient, len(byCategory))
	for categoryCode := range byCategory {
		clients[categoryCode] = m.clients[categoryCode]
	}
	m.mu.RUnlock()

	for categoryCode, items := range byCategory {
		cli := clients[categoryCode]
		if cli != nil && cli.IsLeader() {
			if err := cli.subscribeByClientMessages(items); err != nil {
				return err
			}
		}
	}

	return nil
}

func sha1Hex(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func normalizeUniqueMessages(msgs []server.ClientMessage) []server.ClientMessage {
	out := make([]server.ClientMessage, 0, len(msgs))
	seen := make(map[string]struct{}, len(msgs))
	for _, msg := range msgs {
		msg = server.NormalizeClientMessage(msg)
		key := server.BuildTopicKey(msg)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, msg)
	}
	return out
}
