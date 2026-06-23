package chat_ws

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chat-api/internal/jwt"
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
	eventConnected             = chat.ChatWsEventConnected
	eventError                 = chat.ChatWsEventError
	eventSendUserMessage       = chat.ChatWsEventSendUserMessage
	eventSendUserMessageResult = chat.ChatWsEventSendUserMessageResult
	guestSessionPrefix         = "GS"
	guestMessagePrefix         = "GM"
	guestUsername              = "guest"
	wsProtocol                 = "wklive-chat"
	wsProtocolMerchantPrefix   = "merchant."
	wsProtocolUserPrefix       = "user."
	wsProtocolNicknamePrefix   = "nickname."
	wsProtocolAvatarPrefix     = "avatar."
	successCode                = 200
)

var upgrader = websocket.Upgrader{
	Subprotocols: []string{wsProtocol},
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
	AvatarUrl  string
	SessionNo  string
	Temporary  bool
}

func buildWSIdentity(r *http.Request, svcCtx *svc.ServiceContext) (chatWSIdentity, error) {
	claims, err := jwt.Verify(svcCtx.Config.Jwt.AccessSecret, jwt.TokenFromRequest(r))
	if err != nil {
		return chatWSIdentity{}, err
	}
	username := firstNonEmpty(claims.Nickname, fmt.Sprintf("user-%d", claims.UserId))
	sessionNo, err := openPersistentSession(r.Context(), svcCtx, claims.MerchantId, claims.UserId, username, claims.AvatarUrl)
	if err != nil {
		return chatWSIdentity{}, err
	}
	logx.Infof("chat ws identity resolved by chatToken, merchantId=%d userId=%d sessionNo=%s nickname=%s", claims.MerchantId, claims.UserId, sessionNo, username)
	return chatWSIdentity{
		MerchantId: claims.MerchantId,
		UserId:     claims.UserId,
		Username:   username,
		AvatarUrl:  claims.AvatarUrl,
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
		identity.AvatarUrl,
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
		"nickname":   identity.Username,
		"avatarUrl":  identity.AvatarUrl,
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
		SenderNickname:  firstNonEmpty(conn.Username, data.SenderNickname),
		SenderAvatarUrl: firstNonEmpty(conn.AvatarUrl, data.SenderAvatarUrl),
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
	publishTransientQueueEvent(ctx, svcCtx, conn, "正在排队，客服会尽快接入。")
	conn.SendJSON(eventSendUserMessageResult, &chat.AppChatMessageResp{Base: helper.OkResp(), Data: msg})
}

func sendWSError(conn *ws.Connection, message string) {
	conn.SendJSON(eventError, map[string]string{"message": message})
}

func openPersistentSession(ctx context.Context, svcCtx *svc.ServiceContext, merchantId, userId int64, nickname, avatarUrl string) (string, error) {
	ctx = contextWithChatIdentity(ctx, merchantId, userId)
	resp, err := svcCtx.ChatAppCli.OpenChatSession(ctx, &chat.OpenChatSessionReq{
		Source:          chat.ChatSessionSource_CHAT_SESSION_SOURCE_WEB,
		Title:           strings.TrimSpace(nickname),
		SenderNickname:  strings.TrimSpace(nickname),
		SenderAvatarUrl: strings.TrimSpace(avatarUrl),
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
	logx.Infof("chat ws persistent session opened, merchantId=%d userId=%d sessionNo=%s", merchantId, userId, sessionNo)
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
			AvatarUrl: firstNonEmpty(data.SenderAvatarUrl, conn.AvatarUrl),
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

func publishTransientQueueEvent(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, message string) {
	if svcCtx.BusRedis == nil || conn == nil {
		return
	}
	now := time.Now().UnixMilli()
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatMessageEventTypeQueueUpdated,
		CreatedAt: now,
		Data: &chat.ChatMessage{
			MessageNo:   nextGuestNo(guestMessagePrefix),
			SessionNo:   conn.SessionNo,
			MerchantId:  conn.MerchantId,
			UserId:      conn.UserId,
			SenderType:  chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
			MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
			Content:     message,
			Status:      chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
			CreateTimes: now,
			UpdateTimes: now,
			Sender: &chat.ChatMessageSender{
				Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
				Nickname: "系统",
			},
		},
		Queue: &chat.ChatQueueInfo{
			MerchantId:  conn.MerchantId,
			SessionNo:   conn.SessionNo,
			UserId:      conn.UserId,
			Message:     message,
			UpdateTimes: now,
		},
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		logx.Errorf("marshal guest chat queue event failed: %v", err)
		return
	}
	if _, err := svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload)); err != nil {
		logx.Errorf("publish guest chat queue event failed: %v", err)
	}
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

func merchantIDFromRequest(r *http.Request) (int64, error) {
	raw := strings.TrimSpace(r.Header.Get(utils.CtxKeyMerchantId))
	if raw == "" {
		raw = strings.TrimSpace(wsProtocolValue(r, wsProtocolMerchantPrefix))
	}
	if raw == "" {
		return 0, fmt.Errorf("merchantId is required")
	}
	merchantId, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || merchantId <= 0 {
		return 0, fmt.Errorf("invalid merchantId")
	}
	return merchantId, nil
}

func userIDFromRequest(r *http.Request) (int64, bool, error) {
	raw := strings.TrimSpace(r.Header.Get(utils.CtxKeyUid))
	if raw == "" {
		raw = strings.TrimSpace(wsProtocolValue(r, wsProtocolUserPrefix))
	}
	if raw == "" {
		return 0, false, nil
	}
	userId, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || userId <= 0 {
		return 0, false, fmt.Errorf("invalid userId")
	}
	return userId, true, nil
}

func userNicknameFromRequest(r *http.Request) string {
	return firstNonEmpty(
		r.Header.Get("x-user-nickname"),
		decodeWSProtocolValue(wsProtocolValue(r, wsProtocolNicknamePrefix)),
	)
}

func userAvatarFromRequest(r *http.Request) string {
	return firstNonEmpty(
		r.Header.Get("x-user-avatar-url"),
		decodeWSProtocolValue(wsProtocolValue(r, wsProtocolAvatarPrefix)),
	)
}

func decodeWSProtocolValue(value string) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return ""
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return ""
	}
	return string(decoded)
}

func wsProtocolValue(r *http.Request, prefix string) string {
	for _, part := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		value := strings.TrimSpace(part)
		if strings.HasPrefix(value, prefix) {
			return strings.TrimPrefix(value, prefix)
		}
	}
	return ""
}
