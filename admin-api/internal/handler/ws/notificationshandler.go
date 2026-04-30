package ws

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/ws"
	"wklive/common/notify"
	"wklive/common/utils"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NotificationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := parseToken(r, svcCtx.Config.Jwt.AccessSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("upgrade admin ws failed, userId=%d err=%v", claims.UserId, err)
			return
		}

		client := ws.NewConnection(svcCtx.NotificationHub, conn, claims.UserId, claims.Username)
		svcCtx.NotificationHub.Register(client)
		connected, _ := json.Marshal(notify.Event{
			ID:        "connected",
			Type:      "system",
			Level:     notify.EventLevelInfo,
			Title:     "connected",
			Message:   "admin notification websocket connected",
			UserID:    claims.UserId,
			CreatedAt: time.Now().UnixMilli(),
		})
		client.Send <- connected

		go client.WritePump()
		client.ReadPump()
	}
}

func parseToken(r *http.Request, secret string) (*utils.Claims, error) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("Sec-WebSocket-Protocol"))
	}
	if token == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}

	return utils.ParseToken(secret, token)
}
