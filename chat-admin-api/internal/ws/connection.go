package ws

import (
	"encoding/json"
	"time"

	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type InboundEvent struct {
	EventType chat.ChatEventType `json:"eventType"`
	Data      json.RawMessage    `json:"data"`
}

type Connection struct {
	Conn        *websocket.Conn
	Send        chan []byte
	UserId      int64
	Nickname    string
	AvatarUrl   string
	MerchantId  int64
	AgentId     int64
	AgentUserId int64
	SessionNo   string
	OnMessage   func(*Connection, InboundEvent)
	OnClose     func(*Connection)
	IsGuest     bool
}

func NewConnection(conn *websocket.Conn, nickname string, avatarUrl string, merchantId int64, agentId int64, agentUserId int64, sessionNo string, onMessage func(*Connection, InboundEvent), onClose func(*Connection)) *Connection {
	return &Connection{
		Conn:        conn,
		Send:        make(chan []byte, 32),
		UserId:      -1,
		Nickname:    nickname,
		AvatarUrl:   avatarUrl,
		MerchantId:  merchantId,
		AgentId:     agentId,
		AgentUserId: agentUserId,
		SessionNo:   sessionNo,
		OnMessage:   onMessage,
		OnClose:     onClose,
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
				logx.Errorf("chat admin ws read failed, userId=%d merchantId=%d agentId=%d err=%v", c.UserId, c.MerchantId, c.AgentId, err)
			}
			return
		}
		if len(payload) == 0 || c.OnMessage == nil {
			continue
		}
		var event InboundEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			c.SendEvent(&chat.ChatMessageEvent{
				Code:      200,
				Msg:       "",
				EventType: chat.ChatEventType_CHAT_EVENT_TYPE_ERROR,
				CreatedAt: utils.NowMillis(),
				Payload: &chat.ChatMessageEvent_Error{
					Error: &chat.ChatErrorPayload{
						SessionNo:    c.SessionNo,
						MessageNo:    "",
						ErrorCode:    0,
						ErrorMessage: "invalid json",
						Detail:       err.Error(),
						Retryable:    false,
					},
				},
			})
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
