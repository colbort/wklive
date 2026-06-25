package chat_ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"chat-api/internal/svc"
	"chat-api/internal/types"
	"chat-api/internal/ws"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	guestMessagePrefix = "GM"
	guestUsername      = "guest"
	successCode        = 200
)

type MessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MessagesLogic {
	return &MessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MessagesLogic) Messages(conn *websocket.Conn, req types.ChatWSMessagesReq) {
	if strings.TrimSpace(req.SessionNo) == "" {
		if req.IsGuest {
			logx.Errorf("chat guest ws sessionNo is empty, merchantId=%d userId=%d", req.MerchantId, req.UserId)
			_ = conn.Close()
			return
		}
		sessionNo, err := l.openPersistentSession(l.ctx, req.MerchantId, req.UserId, req.Nickname, req.AvatarUrl)
		if err != nil {
			logx.Errorf("open chat ws persistent session failed, merchantId=%d userId=%d err=%v", req.MerchantId, req.UserId, err)
			_ = conn.Close()
			return
		}
		req.SessionNo = sessionNo
	}

	client := ws.NewConnection(
		l.svcCtx.ChatMessageHub,
		conn,
		req.UserId,
		req.Nickname,
		req.AvatarUrl,
		req.MerchantId,
		req.SessionNo,
		l.onMessage(req.IsGuest),
		l.onClose(),
	)
	l.svcCtx.ChatMessageHub.Register(client)
	// 告诉自己链接成功
	client.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_CONNECTED, connectedPayload(req))
	// TODO 推送（给所有当前商户的坐席）用户进入客服 CHAT_EVENT_TYPE_USER_ONLINE
	// TODO 发送消息给用户（自己）排队信息 CHAT_EVENT_TYPE_QUEUE_INFO
	// if req.IsGuest {
	// 	l.publishTransientQueueEvent(l.ctx, client, "正在排队，客服会尽快接入。")
	// }

	go client.WritePump()
	client.ReadPump()
}

func connectedPayload(req types.ChatWSMessagesReq) map[string]interface{} {
	payload := map[string]interface{}{
		"message":    "chat websocket connected",
		"merchantId": req.MerchantId,
		"userId":     req.UserId,
		"nickname":   req.Nickname,
		"avatarUrl":  req.AvatarUrl,
		"sessionNo":  req.SessionNo,
	}
	if req.IsGuest {
		payload["message"] = "guest chat websocket connected"
		payload["isGuest"] = true
	}
	return payload
}

// 处理用户通过 ws 发送的消息
func (l *MessagesLogic) onMessage(isGuest bool) func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE:
			l.handleSendUserMessage(context.Background(), conn, event.Data, isGuest)
		default:
			sendWSError(conn, "unsupported event type")
		}
	}
}

// 处理用户离开
func (l *MessagesLogic) onClose() func(*ws.Connection) {
	return func(conn *ws.Connection) {
		if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
			return
		}
		now := time.Now().UnixMilli()
		message := "用户已离开客服页面"
		event := &chat.ChatMessageEvent{
			Type:      chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSED,
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
		}
		if err := l.publishChatEvent(context.Background(), event); err != nil {
			logx.Errorf("publish chat user left event failed: %v", err)
		}
	}
}

func (l *MessagesLogic) handleSendUserMessage(ctx context.Context, conn *ws.Connection, payload json.RawMessage, isGuest bool) {
	data, ok := parseSendUserMessagePayload(conn, payload)
	if !ok {
		return
	}
	if strings.TrimSpace(conn.SessionNo) == "" {
		sendWSError(conn, "sessionNo is required")
		return
	}
	if isGuest {
		l.sendTransientUserMessage(ctx, conn, data)
		return
	}
	l.sendPersistentUserMessage(ctx, conn, data)
}

func parseSendUserMessagePayload(conn *ws.Connection, payload json.RawMessage) (UserMessagePayload, bool) {
	var data UserMessagePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		sendWSError(conn, "invalid send_user_message payload")
		return data, false
	}
	return data, true
}

func (l *MessagesLogic) sendPersistentUserMessage(ctx context.Context, conn *ws.Connection, data UserMessagePayload) {
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
	resp, err := l.svcCtx.ChatAppCli.SendUserMessage(contextWithChatIdentity(ctx, conn.MerchantId, conn.UserId), &req)
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_SEND_USER_MESSAGE_RESULT, resp)
}

func (l *MessagesLogic) sendTransientUserMessage(ctx context.Context, conn *ws.Connection, data UserMessagePayload) {
	msg := newTransientMessage(conn, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, conn.UserId, data)
	if err := l.publishTransientMessage(ctx, msg); err != nil {
		sendWSError(conn, err.Error())
		return
	}
	l.publishTransientQueueEvent(ctx, conn, "正在排队，客服会尽快接入。")
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_SEND_USER_MESSAGE_RESULT, &chat.AppChatMessageResp{Base: helper.OkResp(), Data: msg})
}

func sendWSError(conn *ws.Connection, message string) {
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_ERROR, map[string]string{"message": message})
}

func (l *MessagesLogic) openPersistentSession(ctx context.Context, merchantId, userId int64, nickname, avatarUrl string) (string, error) {
	ctx = contextWithChatIdentity(ctx, merchantId, userId)
	resp, err := l.svcCtx.ChatAppCli.OpenChatSession(ctx, &chat.OpenChatSessionReq{
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

func newTransientMessage(conn *ws.Connection, senderType chat.ChatSenderType, senderId int64, data UserMessagePayload) *chat.ChatMessage {
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

func (l *MessagesLogic) publishTransientMessage(ctx context.Context, msg *chat.ChatMessage) error {
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		Data:      msg,
		CreatedAt: time.Now().UnixMilli(),
	}
	return l.publishChatEvent(ctx, event)
}

func (l *MessagesLogic) publishTransientQueueEvent(ctx context.Context, conn *ws.Connection, message string) {
	if conn == nil {
		return
	}
	now := time.Now().UnixMilli()
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_INFO,
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
	if err := l.publishChatEvent(ctx, event); err != nil {
		logx.Errorf("publish guest chat queue event failed: %v", err)
	}
}

func (l *MessagesLogic) publishChatEvent(ctx context.Context, event *chat.ChatMessageEvent) error {
	if l.svcCtx.BusRedis == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		return err
	}
	_, err = l.svcCtx.BusRedis.PublishCtx(ctx, chat.ChatMessageChannel, string(payload))
	return err
}

func nextGuestNo(prefix string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%d%s", prefix, time.Now().UnixMilli(), hex.EncodeToString(b))
}

type UserMessagePayload struct {
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
