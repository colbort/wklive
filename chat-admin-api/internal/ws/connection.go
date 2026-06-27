package ws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type InboundEvent struct {
	Type chat.ChatEventType `json:"type"`
	Data json.RawMessage    `json:"data"`
}

type Connection struct {
	Conn       *websocket.Conn
	Send       chan []byte
	UserId     int64
	Nickname   string
	AvatarUrl  string
	MerchantId int64
	AgentId    int64
	SessionNo  string
	OnMessage  func(*Connection, InboundEvent)
	OnClose    func(*Connection)
	IsGuest    bool
	mu         sync.RWMutex
	session    *chat.ChatSession
}

func NewConnection(conn *websocket.Conn, userId int64, nickname string, avatarUrl string, merchantId, agentId int64, sessionNo string, onMessage func(*Connection, InboundEvent), onClose func(*Connection)) *Connection {
	return &Connection{
		Conn:       conn,
		Send:       make(chan []byte, 32),
		UserId:     userId,
		Nickname:   nickname,
		AvatarUrl:  avatarUrl,
		MerchantId: merchantId,
		AgentId:    agentId,
		SessionNo:  sessionNo,
		OnMessage:  onMessage,
		OnClose:    onClose,
	}
}

func (c *Connection) SetChatSession(session *chat.ChatSession) {
	if c == nil || session == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsGuest = session.GetIsGuest()
	c.session = proto.Clone(session).(*chat.ChatSession)
}

func (c *Connection) ChatSession() *chat.ChatSession {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.session == nil {
		return nil
	}
	return proto.Clone(c.session).(*chat.ChatSession)
}

func (c *Connection) IsGuestSession() bool {
	if c == nil {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsGuest
}

func (c *Connection) ChatSessionUserId(userId int64) int64 {
	if userId != 0 {
		return userId
	}
	session := c.ChatSession()
	if session == nil {
		return 0
	}
	return session.GetUserId()
}

func (c *Connection) ReadPump() {
	defer func() {
		if c.OnClose != nil {
			c.OnClose(c)
		}
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			if isUnexpectedReadClose(err) {
				logx.Errorf("chat admin ws read failed, userId=%d merchantId=%d agentId=%d err=%v", c.UserId, c.MerchantId, c.AgentId, err)
			}
			return
		}
		if len(payload) == 0 || c.OnMessage == nil {
			continue
		}
		event, err := DecodeInboundEvent(payload)
		if err != nil {
			c.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_ERROR, map[string]string{"message": "invalid json"})
			continue
		}
		c.OnMessage(c, event)
	}
}

func isUnexpectedReadClose(err error) bool {
	return websocket.IsUnexpectedCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
	)
}

func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Connection) SendJSON(eventType chat.ChatEventType, data interface{}) {
	payload, err := json.Marshal(map[string]interface{}{
		"type": eventType.String(),
		"data": data,
	})
	if err != nil {
		logx.Errorf("marshal chat ws response failed: %v", err)
		return
	}
	select {
	case c.Send <- payload:
	default:
		logx.Errorf("chat ws send queue is full, userId=%d", c.UserId)
	}
}

func (c *Connection) SendEvent(event *chat.ChatMessageEvent) {
	if event == nil {
		return
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.Errorf("marshal chat admin ws event failed: %v", err)
		return
	}
	select {
	case c.Send <- payload:
	default:
		logx.Errorf("chat admin ws send queue is full, userId=%d", c.UserId)
	}
}

func DecodeInboundEvent(payload []byte) (InboundEvent, error) {
	var raw struct {
		Type      json.RawMessage `json:"type"`
		EventType json.RawMessage `json:"eventType"`
		Data      json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return InboundEvent{}, fmt.Errorf("invalid json")
	}

	typeRaw := raw.Type
	if len(typeRaw) == 0 {
		typeRaw = raw.EventType
	}
	if len(typeRaw) == 0 {
		return InboundEvent{}, fmt.Errorf("event type is required")
	}

	eventType, err := parseChatEventType(typeRaw)
	if err != nil {
		return InboundEvent{}, err
	}
	data := raw.Data
	if len(data) == 0 || string(data) == "null" {
		data = payload
	}
	return InboundEvent{Type: eventType, Data: data}, nil
}

func parseChatEventType(raw json.RawMessage) (chat.ChatEventType, error) {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN, fmt.Errorf("event type is required")
	}

	if value[0] == '"' {
		var name string
		if err := json.Unmarshal(raw, &name); err != nil {
			return chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN, fmt.Errorf("invalid event type")
		}
		return chatEventTypeByName(name)
	}

	n, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN, fmt.Errorf("invalid event type")
	}
	return chat.ChatEventType(n), nil
}

func chatEventTypeByName(name string) (chat.ChatEventType, error) {
	name = strings.TrimSpace(strings.ToUpper(name))
	name = strings.TrimPrefix(name, "CHAT_EVENT_TYPE_")
	fullName := "CHAT_EVENT_TYPE_" + name
	if n, ok := chat.ChatEventType_value[fullName]; ok {
		return chat.ChatEventType(n), nil
	}
	return chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN, fmt.Errorf("unsupported event type: %s", name)
}
