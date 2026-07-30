package ws

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Connection struct {
	Hub           *Hub
	Conn          *websocket.Conn
	Send          chan []byte
	UserId        int64
	Username      string
	TenantId      int64
	IsSystemAdmin bool
	Permissions   map[string]struct{}
	Store         *NotificationStore
}

func NewConnection(
	hub *Hub,
	conn *websocket.Conn,
	userId int64,
	username string,
	tenantId int64,
	isSystemAdmin bool,
	permissions []string,
	store *NotificationStore,
) *Connection {
	permissionSet := make(map[string]struct{}, len(permissions))
	for _, permission := range permissions {
		permissionSet[permission] = struct{}{}
	}
	return &Connection{
		Hub:           hub,
		Conn:          conn,
		Send:          make(chan []byte, 32),
		UserId:        userId,
		Username:      username,
		TenantId:      tenantId,
		IsSystemAdmin: isSystemAdmin,
		Permissions:   permissionSet,
		Store:         store,
	}
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
		messageType, payload, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Errorf("admin ws read failed, userId=%d err=%v", c.UserId, err)
			}
			return
		}
		if messageType != websocket.TextMessage || c.Store == nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		response := c.Store.HandleClientMessage(ctx, c, payload)
		cancel()
		select {
		case c.Send <- response:
		default:
			logx.Errorf("admin ws response queue is full, userId=%d", c.UserId)
			return
		}
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
