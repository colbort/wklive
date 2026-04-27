package ws

import "github.com/zeromicro/go-zero/core/logx"

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
			for client := range h.clients {
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
