package client

import (
	"context"
	"fmt"

	"wklive/services/itick/internal/socket/server"
)

type ItickManager struct {
	clients map[string]*ItickWsClient
}

func NewItickManager(token string, hub *server.Hub) *ItickManager {
	return &ItickManager{
		clients: map[string]*ItickWsClient{
			"crypto":  NewItickWsClient("wss://api.itick.org/crypto", token, hub),
			"forex":   NewItickWsClient("wss://api.itick.org/forex", token, hub),
			"indices": NewItickWsClient("wss://api.itick.org/indices", token, hub),
			"stock":   NewItickWsClient("wss://api.itick.org/stock", token, hub),
			"future":  NewItickWsClient("wss://api.itick.org/future", token, hub),
			"fund":    NewItickWsClient("wss://api.itick.org/fund", token, hub),
		},
	}
}

func (m *ItickManager) Start(ctx context.Context) {
	for _, cli := range m.clients {
		cli.Start(ctx)
	}
}

func (m *ItickManager) Subscribe(msg server.ClientMessage) error {
	cli, ok := m.clients[msg.Market]
	if !ok {
		return fmt.Errorf("unsupported market: %s", msg.Market)
	}
	return cli.SubscribeByClientMessage(msg)
}

func (m *ItickManager) Unsubscribe(msg server.ClientMessage) error {
	cli, ok := m.clients[msg.Market]
	if !ok {
		return fmt.Errorf("unsupported market: %s", msg.Market)
	}
	return cli.UnsubscribeByClientMessage(msg)
}
