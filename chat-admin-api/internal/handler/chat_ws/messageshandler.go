package chat_ws

import (
	"net/http"
	"strconv"
	"strings"

	"chat-admin-api/internal/logic/chat_ws"
	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"
	"wklive/common/utils"
)

func MessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := parseToken(r, svcCtx.Config.Jwt.AccessSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		l := chat_ws.NewMessagesLogic(r.Context(), svcCtx)
		l.Messages(w, r, types.ChatAdminWSMessagesReq{
			UserId:     claims.UserId,
			MerchantId: parseInt64(r.URL.Query().Get("merchantId")),
			AgentId:    parseInt64(r.URL.Query().Get("agentId")),
		})
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

func parseInt64(value string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return n
}
