package ws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
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
	Hub        *Hub
	Conn       *websocket.Conn
	Send       chan []byte
	UserId     int64
	Nickname   string
	AvatarUrl  string
	MerchantId int64
	AgentId    int64
	SessionNo  string
	OnMessage  func(*Connection, InboundEvent)
}

func NewConnection(hub *Hub, conn *websocket.Conn, userId int64, nickname string, avatarUrl string, merchantId, agentId int64, sessionNo string, onMessage func(*Connection, InboundEvent)) *Connection {
	return &Connection{
		Hub:        hub,
		Conn:       conn,
		Send:       make(chan []byte, 32),
		UserId:     userId,
		Nickname:   nickname,
		AvatarUrl:  avatarUrl,
		MerchantId: merchantId,
		AgentId:    agentId,
		SessionNo:  sessionNo,
		OnMessage:  onMessage,
	}
}

func (c *Connection) Match(message *chat.ChatMessage) bool {
	if message == nil {
		return false
	}
	if c.SessionNo != "" && message.SessionNo != c.SessionNo {
		return false
	}
	agentId := int64FromString(message.GetAgentId())
	if c.AgentId > 0 && agentId != c.AgentId && agentId != 0 {
		return false
	}
	return true
}

func (c *Connection) MatchEvent(event *chat.ChatMessageEvent) bool {
	if event == nil {
		return false
	}
	if event.GetSessionEvent() != nil {
		return c.matchSessionEvent(event.GetSessionEvent())
	}
	if event.GetSession() != nil {
		return c.matchSession(event.GetSession())
	}
	if event.GetQueue() != nil {
		return c.matchQueue(event.GetQueue())
	}
	if event.GetAgent() != nil {
		return c.matchAgent(event.GetAgent())
	}
	if event.GetData() != nil {
		return c.Match(event.GetData())
	}
	return false
}

func (c *Connection) matchSession(session *chat.ChatSession) bool {
	if session == nil {
		return false
	}
	if c.MerchantId > 0 && session.MerchantId != c.MerchantId {
		return false
	}
	if c.SessionNo != "" && session.SessionNo != c.SessionNo {
		return false
	}
	if c.AgentId > 0 && session.AgentId != c.AgentId && session.AgentId != 0 {
		return false
	}
	return true
}

func (c *Connection) matchSessionEvent(event *chat.ChatSessionEvent) bool {
	if event == nil {
		return false
	}
	if c.MerchantId > 0 && event.MerchantId != c.MerchantId {
		return false
	}
	if c.SessionNo != "" && event.SessionNo != c.SessionNo {
		return false
	}
	if c.AgentId > 0 && event.AgentId != c.AgentId && event.AgentId != 0 {
		return false
	}
	return true
}

func (c *Connection) matchQueue(queue *chat.ChatQueueInfo) bool {
	if queue == nil {
		return false
	}
	if c.MerchantId > 0 && queue.MerchantId != c.MerchantId {
		return false
	}
	if c.SessionNo != "" && queue.SessionNo != c.SessionNo {
		return false
	}
	return true
}

func (c *Connection) matchAgent(agent *chat.ChatAgent) bool {
	if agent == nil {
		return false
	}
	if c.MerchantId > 0 && agent.MerchantId != c.MerchantId {
		return false
	}
	if c.AgentId > 0 && agent.Id != c.AgentId {
		return false
	}
	return true
}

func (c *Connection) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
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

func int64FromString(value string) int64 {
	id, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return id
}
