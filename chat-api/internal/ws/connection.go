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
	"google.golang.org/protobuf/encoding/protojson"
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

func NewConnection(conn *websocket.Conn, userId int64, username string, avatarUrl string, merchantId int64, sessionNo string, onMessage func(*Connection, InboundEvent), onClose func(*Connection)) *Connection {
	return &Connection{
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
				logx.Errorf("chat user ws read failed, userId=%d merchantId=%d err=%v", c.UserId, c.MerchantId, err)
			}
			return
		}
		if len(payload) == 0 || c.OnMessage == nil {
			continue
		}
		event, err := DecodeInboundEvent(payload)
		if err != nil {
			c.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_ERROR, map[string]string{"message": err.Error()})
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

// SendJSON writes a chat event envelope to the websocket.
func (c *Connection) SendJSON(eventType chat.ChatEventType, data interface{}) {
	payload, err := json.Marshal(map[string]interface{}{
		"type":     eventType.String(),
		"typeName": eventType.String(),
		"data":     data,
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

func (c *Connection) SendEvent(event *chat.ChatMessageEvent) {
	if event == nil {
		return
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.Errorf("marshal chat user ws event failed: %v", err)
		return
	}
	select {
	case c.Send <- payload:
	default:
		logx.Errorf("chat user ws send queue is full, userId=%d", c.UserId)
	}
}

// DecodeInboundEvent 支持前端用数字或枚举字符串发送事件：
// {"type":1,"data":{...}}
// {"type":"CHAT_EVENT_TYPE_MESSAGE","data":{...}}
// {"eventType":"CHAT_EVENT_TYPE_SESSION_CLOSE","data":{...}}
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
		return InboundEvent{}, fmt.Errorf("event data is required")
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
