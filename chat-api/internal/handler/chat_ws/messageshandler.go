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
	guestMessagePrefix         = "GM"
	guestUsername              = "guest"
	successCode                = 200
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, err := buildWSIdentity(r, svcCtx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		serveWSConnection(w, r, svcCtx, identity)
	}
}

type chatWSIdentity struct {
	MerchantId int64
	UserId     int64
	Username   string
	SessionNo  string
	Temporary  bool
}

func buildWSIdentity(r *http.Request, svcCtx *svc.ServiceContext) (chatWSIdentity, error) {
	merchantId, err := merchantIDFromRequest(r)
	if err != nil {
		return chatWSIdentity{}, err
	}

	sessionNo := nextGuestNo(guestSessionPrefix)

	token, hasToken, err := tokenFromRequest(r)
	if err != nil {
		return chatWSIdentity{
			MerchantId: merchantId,
			SessionNo:  sessionNo,
		}, err
	}
	if !hasToken {
		return chatWSIdentity{
			MerchantId: merchantId,
			UserId:     guestUserID(sessionNo),
			Username:   guestUsername,
			SessionNo:  sessionNo,
			Temporary:  true,
		}, nil
	}

	claims, err := utils.ParseToken(svcCtx.Config.Jwt.AccessSecret, token)
	if err != nil {
		return chatWSIdentity{}, fmt.Errorf("Unauthorized")
	}
	sessionNo, err = openPersistentSession(r.Context(), svcCtx, merchantId, claims.UserId)
	if err != nil {
		return chatWSIdentity{}, err
	}
	return chatWSIdentity{
		MerchantId: merchantId,
		UserId:     claims.UserId,
		Username:   claims.Username,
		SessionNo:  sessionNo,
	}, nil
}

func serveWSConnection(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext, identity chatWSIdentity) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("upgrade chat user ws failed, userId=%d merchantId=%d temporary=%t err=%v", identity.UserId, identity.MerchantId, identity.Temporary, err)
		return
	}

	client := ws.NewConnection(
		svcCtx.ChatMessageHub,
		conn,
		identity.UserId,
		identity.Username,
		identity.MerchantId,
		identity.SessionNo,
		handleInbound(svcCtx, identity.Temporary),
	)
	svcCtx.ChatMessageHub.Register(client)
	client.SendJSON(eventConnected, connectedPayload(identity))

	go client.WritePump()
	client.ReadPump()
}

func connectedPayload(identity chatWSIdentity) map[string]interface{} {
	payload := map[string]interface{}{
		"message":    "chat websocket connected",
		"merchantId": identity.MerchantId,
		"userId":     identity.UserId,
		"sessionNo":  identity.SessionNo,
	}
	if identity.Temporary {
		payload["message"] = "guest chat websocket connected"
		payload["temporary"] = true
	}
	return payload
}

func handleInbound(svcCtx *svc.ServiceContext, temporary bool) func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case eventSendUserMessage:
			handleSendUserMessage(context.Background(), svcCtx, conn, event.Data, temporary)
		default:
			sendWSError(conn, "unsupported event type")
		}
	}
}

func handleSendUserMessage(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, payload json.RawMessage, temporary bool) {
	data, ok := parseSendUserMessagePayload(conn, payload)
	if !ok {
		return
	}
	if strings.TrimSpace(conn.SessionNo) == "" {
		sendWSError(conn, "sessionNo is required")
		return
	}
	if temporary {
		sendTransientUserMessage(ctx, svcCtx, conn, data)
		return
	}
	sendPersistentUserMessage(ctx, svcCtx, conn, data)
}

func parseSendUserMessagePayload(conn *ws.Connection, payload json.RawMessage) (sendUserMessagePayload, bool) {
	var data sendUserMessagePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		sendWSError(conn, "invalid send_user_message payload")
		return data, false
	}
	return data, true
}

func sendPersistentUserMessage(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, data sendUserMessagePayload) {
	req := chat.SendUserMessageReq{
		SessionNo:       conn.SessionNo,
		MessageType:     chat.ChatMessageType(data.MessageType),
		Content:         strings.TrimSpace(data.Content),
		MediaUrl:        strings.TrimSpace(data.MediaUrl),
		MediaName:       strings.TrimSpace(data.MediaName),
		MediaMime:       strings.TrimSpace(data.MediaMime),
		MediaSize:       data.MediaSize,
		SenderNickname:  firstNonEmpty(data.SenderNickname, conn.Username),
		SenderAvatarUrl: strings.TrimSpace(data.SenderAvatarUrl),
	}
	resp, err := svcCtx.ChatAppCli.SendUserMessage(contextWithChatIdentity(ctx, conn.MerchantId, conn.UserId), &req)
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(eventSendUserMessageResult, resp)
}

func sendTransientUserMessage(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, data sendUserMessagePayload) {
	msg := newTransientMessage(conn, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, conn.UserId, data)
	if err := publishTransientMessage(ctx, svcCtx, msg); err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(eventSendUserMessageResult, &chat.AppChatMessageResp{Base: helper.OkResp(), Data: msg})
}

func sendWSError(conn *ws.Connection, message string) {
	conn.SendJSON(eventError, map[string]string{"message": message})
}

func openPersistentSession(ctx context.Context, svcCtx *svc.ServiceContext, merchantId, userId int64) (string, error) {
	ctx = contextWithChatIdentity(ctx, merchantId, userId)
	resp, err := svcCtx.ChatAppCli.OpenChatSession(ctx, &chat.OpenChatSessionReq{
		Source: chat.ChatSessionSource_CHAT_SESSION_SOURCE_WEB,
	})
	if err != nil {
		return "", err
	}
	if resp.GetBase().GetCode() != successCode {
		return "", fmt.Errorf("%s", resp.GetBase().GetMsg())
	}
	sessionNo := strings.TrimSpace(resp.GetData().GetSessionNo())
	if sessionNo == "" {
		return "", fmt.Errorf("sessionNo is empty")
	}
	return sessionNo, nil
}

func contextWithChatIdentity(ctx context.Context, merchantId, userId int64) context.Context {
	ctx = context.WithValue(ctx, utils.CtxKeyMerchantId, merchantId)
	ctx = context.WithValue(ctx, utils.CtxKeyUid, userId)
	return ctx
}

func newTransientMessage(conn *ws.Connection, senderType chat.ChatSenderType, senderId int64, data sendUserMessagePayload) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	senderNickname := firstNonEmpty(data.SenderNickname, conn.Username, guestUsername)
	return &chat.ChatMessage{
		MessageNo:  nextGuestNo(guestMessagePrefix),
		SessionNo:  conn.SessionNo,
		MerchantId: conn.MerchantId,
		UserId:     conn.UserId,
		AgentId:    0,
		SenderType: senderType,
		Sender: &chat.ChatMessageSender{
			Id:        senderId,
			Type:      senderType,
			Nickname:  senderNickname,
			AvatarUrl: strings.TrimSpace(data.SenderAvatarUrl),
		},
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

type sendUserMessagePayload struct {
	MessageType     int64  `json:"messageType"`
	Content         string `json:"content"`
	MediaUrl        string `json:"mediaUrl"`
	MediaName       string `json:"mediaName"`
	MediaMime       string `json:"mediaMime"`
	MediaSize       int64  `json:"mediaSize"`
	SenderNickname  string `json:"senderNickname"`
	SenderAvatarUrl string `json:"senderAvatarUrl"`
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}

func tokenFromRequest(r *http.Request) (string, bool, error) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return "", false, nil
	}
	fields := strings.Fields(auth)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") || strings.TrimSpace(fields[1]) == "" {
		return "", false, fmt.Errorf("invalid Authorization header")
	}
	return fields[1], true, nil
}

func merchantIDFromRequest(r *http.Request) (int64, error) {
	raw := strings.TrimSpace(r.Header.Get(utils.CtxKeyMerchantId))
	if raw == "" {
		return 0, fmt.Errorf("merchantId is required")
	}
	merchantId, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || merchantId <= 0 {
		return 0, fmt.Errorf("invalid merchantId")
	}
	return merchantId, nil
}
