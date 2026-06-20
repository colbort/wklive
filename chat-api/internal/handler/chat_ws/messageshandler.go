package chat_ws

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chat-api/internal/svc"
	"chat-api/internal/ws"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	eventConnected             = "connected"
	eventError                 = "error"
	eventSendUserMessage       = "send_user_message"
	eventSendUserMessageResult = "send_user_message.result"
	guestSessionPrefix         = "GS"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := tokenFromRequest(r)
		if token == "" {
			serveGuestMessages(w, r, svcCtx)
			return
		}

		claims, err := utils.ParseToken(svcCtx.Config.Jwt.AccessSecret, token)
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

func serveGuestMessages(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext) {
	merchantId := parseInt64(r.URL.Query().Get("merchantId"))
	if merchantId <= 0 {
		http.Error(w, "merchantId is required", http.StatusBadRequest)
		return
	}

	sessionNo := strings.TrimSpace(r.URL.Query().Get("sessionNo"))
	if sessionNo == "" {
		sessionNo = nextGuestNo(guestSessionPrefix)
	} else if !isGuestSession(sessionNo) {
		http.Error(w, "invalid guest sessionNo", http.StatusBadRequest)
		return
	}
	userId := guestUserID(sessionNo)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := ws.NewConnection(svcCtx.ChatMessageHub, conn, userId, "guest", merchantId, sessionNo, handleGuestInbound(svcCtx))
	svcCtx.ChatMessageHub.Register(client)
	client.SendJSON(eventConnected, map[string]interface{}{
		"message":    "guest chat websocket connected",
		"merchantId": merchantId,
		"userId":     userId,
		"sessionNo":  sessionNo,
		"temporary":  true,
	})

	go client.WritePump()
	client.ReadPump()
}

func handleGuestInbound(svcCtx *svc.ServiceContext) func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case eventSendUserMessage:
			handleSendGuestUserMessage(context.Background(), svcCtx, conn, event.Data)
		default:
			conn.SendJSON(eventError, map[string]string{"message": "unsupported event type"})
		}
	}
}

func handleSendGuestUserMessage(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, payload json.RawMessage) {
	var data sendUserMessagePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		conn.SendJSON(eventError, map[string]string{"message": "invalid send_user_message payload"})
		return
	}
	msg := newTransientMessage(conn.MerchantId, conn.SessionNo, conn.UserId, 0, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, conn.UserId, data)
	if err := publishTransientMessage(ctx, svcCtx, msg); err != nil {
		conn.SendJSON(eventError, map[string]string{"message": err.Error()})
		return
	}
	conn.SendJSON(eventSendUserMessageResult, &chat.AppChatMessageResp{Base: helper.OkResp(), Data: msg})
}

func newTransientMessage(merchantId int64, sessionNo string, userId int64, agentId int64, senderType chat.ChatSenderType, senderId int64, data sendUserMessagePayload) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	return &chat.ChatMessage{
		MessageNo:   nextGuestNo("GM"),
		SessionNo:   sessionNo,
		MerchantId:  merchantId,
		UserId:      userId,
		AgentId:     agentId,
		SenderType:  senderType,
		SenderId:    senderId,
		MessageType: chat.ChatMessageType(data.MessageType),
		Content:     strings.TrimSpace(data.Content),
		MediaUrl:    strings.TrimSpace(data.MediaUrl),
		MediaName:   strings.TrimSpace(data.MediaName),
		MediaMime:   strings.TrimSpace(data.MediaMime),
		MediaSize:   data.MediaSize,
		Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		CreateTimes: now,
		UpdateTimes: now,
	}
}

func publishTransientMessage(ctx context.Context, svcCtx *svc.ServiceContext, msg *chat.ChatMessage) error {
	if svcCtx.BusRedis == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatMessageEventTypeMessage,
		Data:      msg,
		CreatedAt: time.Now().UnixMilli(),
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		return err
	}
	_, err = svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload))
	return err
}

func nextGuestNo(prefix string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%d%s", prefix, time.Now().UnixMilli(), hex.EncodeToString(b))
}

func guestUserID(sessionNo string) int64 {
	sum := sha256.Sum256([]byte(sessionNo))
	n := int64(binary.BigEndian.Uint64(sum[:8]) & 0x3fffffffffffffff)
	if n == 0 {
		return -1
	}
	return -n
}

func isGuestSession(sessionNo string) bool {
	return strings.HasPrefix(strings.TrimSpace(sessionNo), guestSessionPrefix)
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
	return utils.ParseToken(secret, tokenFromRequest(r))
}

func tokenFromRequest(r *http.Request) string {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("Sec-WebSocket-Protocol"))
	}
	if token == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return token
}

func parseInt64(value string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return n
}
