package ws

import (
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
)

type Hub struct {
	register   chan *Connection
	unregister chan *Connection
	broadcast  chan *chat.ChatMessageEvent
	clients    map[*Connection]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Connection),
		unregister: make(chan *Connection),
		broadcast:  make(chan *chat.ChatMessageEvent, 256),
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
		case event := <-h.broadcast:
			payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
			if err != nil {
				logx.Errorf("marshal chat ws event failed: %v", err)
				continue
			}
			for client := range h.clients {
				if !client.Match(event.GetData()) {
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

func (h *Hub) Broadcast(event *chat.ChatMessageEvent) {
	if event == nil {
		return
	}
	select {
	case h.broadcast <- event:
	default:
		logx.Errorf("chat ws broadcast queue is full, drop message event")
	}
}

func (h *Hub) Register(client *Connection) {
	h.register <- client
}

func (h *Hub) Unregister(client *Connection) {
	h.unregister <- client
}
