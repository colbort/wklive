package chat_ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/ws"
	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	eventConnected              = "connected"
	eventError                  = "error"
	eventSendAgentMessage       = "send_agent_message"
	eventSendAgentMessageResult = "send_agent_message.result"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := parseToken(r, svcCtx.Config.Jwt.AccessSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		merchantId := parseInt64(r.URL.Query().Get("merchantId"))
		agentId := parseInt64(r.URL.Query().Get("agentId"))
		sessionNo := strings.TrimSpace(r.URL.Query().Get("sessionNo"))

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("upgrade chat admin ws failed, userId=%d err=%v", claims.UserId, err)
			return
		}

		client := ws.NewConnection(svcCtx.ChatMessageHub, conn, claims.UserId, claims.Username, merchantId, agentId, sessionNo, handleInbound(svcCtx))
		svcCtx.ChatMessageHub.Register(client)
		client.SendJSON(eventConnected, map[string]interface{}{
			"message":    "chat admin websocket connected",
			"merchantId": merchantId,
			"agentId":    agentId,
			"sessionNo":  sessionNo,
		})

		go client.WritePump()
		client.ReadPump()
	}
}

func handleInbound(svcCtx *svc.ServiceContext) func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case eventSendAgentMessage:
			handleSendAgentMessage(context.Background(), svcCtx, conn, event.Data)
		default:
			conn.SendJSON(eventError, map[string]string{"message": "unsupported event type"})
		}
	}
}

func handleSendAgentMessage(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, payload json.RawMessage) {
	var data sendAgentMessagePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		conn.SendJSON(eventError, map[string]string{"message": "invalid send_agent_message payload"})
		return
	}
	req := chat.SendAgentMessageReq{
		MerchantId:  data.MerchantId,
		AgentId:     data.AgentId,
		SessionNo:   data.SessionNo,
		MessageType: chat.ChatMessageType(data.MessageType),
		Content:     data.Content,
		MediaUrl:    data.MediaUrl,
		MediaName:   data.MediaName,
		MediaMime:   data.MediaMime,
		MediaSize:   data.MediaSize,
	}
	if req.MerchantId == 0 {
		req.MerchantId = conn.MerchantId
	}
	if req.AgentId == 0 {
		req.AgentId = conn.AgentId
	}
	if req.SessionNo == "" {
		req.SessionNo = conn.SessionNo
	}

	resp, err := svcCtx.ChatAdminCli.SendAgentMessage(ctx, &req)
	if err != nil {
		conn.SendJSON(eventError, map[string]string{"message": err.Error()})
		return
	}
	conn.SendJSON(eventSendAgentMessageResult, resp)
}

type sendAgentMessagePayload struct {
	MerchantId  int64  `json:"merchantId"`
	AgentId     int64  `json:"agentId"`
	SessionNo   string `json:"sessionNo"`
	MessageType int64  `json:"messageType"`
	Content     string `json:"content"`
	MediaUrl    string `json:"mediaUrl"`
	MediaName   string `json:"mediaName"`
	MediaMime   string `json:"mediaMime"`
	MediaSize   int64  `json:"mediaSize"`
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
