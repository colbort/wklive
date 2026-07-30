package ws

import (
	"encoding/json"

	"wklive/common/notify"

	"github.com/zeromicro/go-zero/core/logx"
)

var eventPermissions = map[string]string{
	notify.EventTypeUserIdentitySubmit: "users:user:identities:list",
	notify.EventTypeRecharge:           "payment:recharge-order:list",
	notify.EventTypeWithdraw:           "payment:withdraw-order:list",
	"contract_reconciliation":          "trade:operation:reconciliation-issue:list",
	"price_engine_input":               "market:price-formula:list",
	"snapshot_outbox":                  "market:snapshot-outbox:list",
}

type Hub struct {
	register   chan *Connection
	unregister chan *Connection
	broadcast  chan []byte
	clients    map[*Connection]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Connection),
		unregister: make(chan *Connection),
		broadcast:  make(chan []byte, 256),
		clients:    make(map[*Connection]struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
		case payload := <-h.broadcast:
			var event notify.Event
			if err := json.Unmarshal(payload, &event); err != nil || event.ID == "" || event.Type == "" {
				logx.Errorf("drop invalid admin ws event: %v", err)
				continue
			}
			for client := range h.clients {
				if !client.CanReceive(event) {
					continue
				}
				select {
				case client.Send <- payload:
				default:
					delete(h.clients, client)
					close(client.Send)
				}
			}
		}
	}
}

func (c *Connection) CanReceive(event notify.Event) bool {
	if c == nil {
		return false
	}
	if c.IsSystemAdmin {
		return true
	}
	if event.TenantID == 0 && isOperationalEvent(event.Type) {
		return false
	}
	if event.TenantID != 0 && event.TenantID != c.TenantId {
		return false
	}
	requiredPermission, known := eventPermissions[event.Type]
	if !known {
		return false
	}
	_, allowed := c.Permissions[requiredPermission]
	return allowed
}

func (h *Hub) BroadcastRaw(payload []byte) {
	select {
	case h.broadcast <- payload:
	default:
		logx.Errorf("admin ws broadcast queue is full, drop raw event")
	}
}

func (h *Hub) Register(client *Connection) {
	h.register <- client
}

func (h *Hub) Unregister(client *Connection) {
	h.unregister <- client
}
