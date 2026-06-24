package ws

import (
	"encoding/json"
	"time"

	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 20 * time.Second
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
	Username   string
	AvatarUrl  string
	MerchantId int64
	SessionNo  string
	OnMessage  func(*Connection, InboundEvent)
	OnClose    func(*Connection)
}

func NewConnection(hub *Hub, conn *websocket.Conn, userId int64, username string, avatarUrl string, merchantId int64, sessionNo string, onMessage func(*Connection, InboundEvent), onClose func(*Connection)) *Connection {
	return &Connection{
		Hub:        hub,
		Conn:       conn,
		Send:       make(chan []byte, 32),
		UserId:     userId,
		Username:   username,
		AvatarUrl:  avatarUrl,
		MerchantId: merchantId,
		SessionNo:  sessionNo,
		OnMessage:  onMessage,
		OnClose:    onClose,
	}
}

func (c *Connection) Match(message *chat.ChatMessage) bool {
	if message == nil {
		return false
	}
	if c.MerchantId > 0 && message.MerchantId != c.MerchantId {
		return false
	}
	if c.SessionNo != "" && message.SessionNo != c.SessionNo {
		return false
	}
	if c.SessionNo == "" && c.UserId > 0 && message.UserId != c.UserId {
		return false
	}
	return true
}

func (c *Connection) MatchEvent(event *chat.ChatMessageEvent) bool {
	if event == nil {
		return false
	}
	if event.GetData() != nil {
		return c.Match(event.GetData())
	}
	if event.GetSession() != nil {
		return c.matchSession(event.GetSession())
	}
	if event.GetSessionEvent() != nil {
		return c.matchSessionEvent(event.GetSessionEvent())
	}
	if event.GetQueue() != nil {
		return c.matchQueue(event.GetQueue())
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
	if c.SessionNo == "" && c.UserId > 0 && session.UserId != c.UserId {
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
	if c.SessionNo == "" && c.UserId > 0 && event.UserId != c.UserId {
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
	if c.SessionNo == "" && c.UserId > 0 && queue.UserId != c.UserId {
		return false
	}
	return true
}

func (c *Connection) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Errorf("chat user ws read failed, userId=%d merchantId=%d err=%v", c.UserId, c.MerchantId, err)
			}
			return
		}
		if len(payload) == 0 || c.OnMessage == nil {
			continue
		}
		var event InboundEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			c.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_ERROR, map[string]string{"message": "invalid json"})
			continue
		}
		c.OnMessage(c, event)
	}
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
		"type": eventType,
		"data": data,
	})
	if err != nil {
		logx.Errorf("marshal chat user ws response failed: %v", err)
		return
	}
	select {
	case c.Send <- payload:
	default:
		logx.Errorf("chat user ws send queue is full, userId=%d", c.UserId)
	}
}
