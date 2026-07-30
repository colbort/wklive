package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/ws"
	"wklive/common/notify"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const notificationSubprotocol = "wklive-admin-notifications"

func NotificationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		originConfig := ws.Config{
			AllowedOrigins:     svcCtx.Config.NotificationWS.AllowedOrigins,
			AllowMissingOrigin: svcCtx.Config.NotificationWS.AllowMissingOrigin,
		}
		if !ws.OriginAllowed(r, originConfig) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		claims, err := parseToken(r, svcCtx.Config.Jwt.AccessSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userResp, err := svcCtx.SystemCli.SysUserDetail(r.Context(), &system.SysUserDetailReq{
			Id: claims.UserId,
		})
		if err != nil || userResp == nil || userResp.Data == nil ||
			userResp.Data.Enabled != common.Enable_ENABLE_ENABLED {
			logx.Errorf("reject admin ws user, userId=%d err=%v", claims.UserId, err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		permResp, err := svcCtx.SystemCli.LoginUserPerms(r.Context(), &system.LoginUserPermsReq{
			UserId: claims.UserId,
		})
		if err != nil || permResp == nil {
			logx.Errorf("load admin ws permissions failed, userId=%d err=%v", claims.UserId, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		upgrader := websocket.Upgrader{
			Subprotocols: []string{notificationSubprotocol},
			CheckOrigin: func(request *http.Request) bool {
				return ws.OriginAllowed(request, originConfig)
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("upgrade admin ws failed, userId=%d err=%v", claims.UserId, err)
			return
		}

		client := ws.NewConnection(
			svcCtx.NotificationHub,
			conn,
			claims.UserId,
			claims.Username,
			userResp.Data.TenantId,
			userResp.Data.UserType == system.UserType_USER_TYPE_SYSTEM_ADMIN,
			permResp.Perms,
			svcCtx.NotificationStore,
		)
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
	if strings.TrimSpace(r.URL.Query().Get("token")) != "" {
		return nil, errors.New("websocket URL token is not allowed")
	}
	var token string
	protocolFound := false
	for _, protocol := range websocket.Subprotocols(r) {
		switch {
		case protocol == notificationSubprotocol:
			protocolFound = true
		case strings.HasPrefix(protocol, "bearer."):
			token = strings.TrimSpace(strings.TrimPrefix(protocol, "bearer."))
		}
	}
	if !protocolFound {
		return nil, errors.New("websocket notification subprotocol is required")
	}
	if token == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	if token == "" {
		return nil, errors.New("websocket bearer token is required")
	}

	return utils.ParseToken(secret, token)
}
