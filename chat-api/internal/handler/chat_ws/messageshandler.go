package chat_ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chat-api/internal/svc"
	"chat-api/internal/ws"
	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	eventConnected             = "connected"
	eventError                 = "error"
	eventSendUserMessage       = "send_user_message"
	eventSendUserMessageResult = "send_user_message.result"
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
		sessionNo := strings.TrimSpace(r.URL.Query().Get("sessionNo"))

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("upgrade chat user ws failed, userId=%d err=%v", claims.UserId, err)
			return
		}

		client := ws.NewConnection(svcCtx.ChatMessageHub, conn, claims.UserId, claims.Username, merchantId, sessionNo, handleInbound(svcCtx))
		svcCtx.ChatMessageHub.Register(client)
		client.SendJSON(eventConnected, map[string]interface{}{
			"message":    "chat websocket connected",
			"merchantId": merchantId,
			"userId":     claims.UserId,
			"sessionNo":  sessionNo,
		})

		go client.WritePump()
		client.ReadPump()
	}
}

func handleInbound(svcCtx *svc.ServiceContext) func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case eventSendUserMessage:
			handleSendUserMessage(context.Background(), svcCtx, conn, event.Data)
		default:
			conn.SendJSON(eventError, map[string]string{"message": "unsupported event type"})
		}
	}
}

func handleSendUserMessage(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, payload json.RawMessage) {
	var data sendUserMessagePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		conn.SendJSON(eventError, map[string]string{"message": "invalid send_user_message payload"})
		return
	}
	req := chat.SendUserMessageReq{
		MerchantId:  data.MerchantId,
		UserId:      conn.UserId,
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
	if req.SessionNo == "" {
		req.SessionNo = conn.SessionNo
	}

	resp, err := svcCtx.ChatAppCli.SendUserMessage(ctx, &req)
	if err != nil {
		conn.SendJSON(eventError, map[string]string{"message": err.Error()})
		return
	}
	conn.SendJSON(eventSendUserMessageResult, resp)
}

type sendUserMessagePayload struct {
	MerchantId  int64  `json:"merchantId"`
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
