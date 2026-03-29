package client

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"wklive/services/itick/internal/socket/server"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ItickManager struct {
	wsurl string
	token string
	hub   *server.Hub
	model models.ItickCategoryModel

	mu      sync.RWMutex
	clients map[string]*ItickWsClient
}

func NewItickManager(wsurl string, token string, hub *server.Hub, model models.ItickCategoryModel) *ItickManager {
	return &ItickManager{
		wsurl:   wsurl,
		token:   token,
		hub:     hub,
		model:   model,
		clients: make(map[string]*ItickWsClient),
	}
}

func (m *ItickManager) Load(ctx context.Context) error {
	// 这里改成你自己的 model 查询方法
	categories, err := m.model.FindAll(ctx)
	if err != nil {
		return err
	}

	newClients := make(map[string]*ItickWsClient)

	for _, item := range categories {
		market := strings.ToLower(strings.TrimSpace(item.CategoryCode))
		wsURL := strings.TrimSpace(m.wsurl)

		if market == "" || wsURL == "" {
			logx.Errorf("skip invalid itick category, code=%s, wsURL=%s", item.CategoryCode, m.wsurl)
			continue
		}

		newClients[market] = NewItickWsClient(fmt.Sprintf("%s/%s", wsURL, market), m.token, market, m.hub)
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

func (m *ItickManager) Start(ctx context.Context) {
	m.mu.RLock()
	clients := make([]*ItickWsClient, 0, len(m.clients))
	for _, cli := range m.clients {
		clients = append(clients, cli)
	}
	m.mu.RUnlock()

	for _, cli := range clients {
		cli.Start(ctx)
	}
}

func (m *ItickManager) Subscribe(msg server.ClientMessage) error {
	m.mu.RLock()
	cli, ok := m.clients[msg.Market]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unsupported market: %s", msg.Market)
	}
	return cli.SubscribeByClientMessage(msg)
}

func (m *ItickManager) Unsubscribe(msg server.ClientMessage) error {
	m.mu.RLock()
	cli, ok := m.clients[msg.Market]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unsupported market: %s", msg.Market)
	}
	return cli.UnsubscribeByClientMessage(msg)
}
