package chat_ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"chat-api/internal/svc"
	"chat-api/internal/types"
	"chat-api/internal/ws"
	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	guestMessagePrefix = "GM"
	guestUsername      = "guest"
	systemNickname     = "系统"
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

// Messages 是 chat-api 用户侧 websocket 入口。
// 流程：
// 1. chat-ui 建立 ws 连接并携带 token；
// 2. chat-api 返回连接成功和当前排队信息；
// 3. chat-api 发布 USER_JOIN 给 chat-admin-api，由 chat-admin-api 转发给所有坐席；
// 4. 后续 chat-admin-api 发布 AGENT_ASSIGNED / MESSAGE / SESSION_CLOSE 等事件时，由 subscriber -> hub 转发给匹配的 chat-ui。
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
		l.onClose(req.IsGuest),
	)
	l.svcCtx.ChatMessageHub.Register(client)

	// 先给 chat-ui 返回连接成功和排队信息。
	l.sendConnectedEvent(client)
	l.sendQueueEvent(client, chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_JOIN, "正在排队，客服会尽快接入。")

	// 再通知后台：用户上线，进入待接待列表。
	l.publishUserOnlineEvent(l.ctx, client, req.IsGuest)

	go client.WritePump()
	client.ReadPump()
}

// 处理用户通过 ws 发送的事件。
func (l *MessagesLogic) onMessage(isGuest bool) func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE:
			l.handleSendUserMessage(context.Background(), conn, event.Data, isGuest)
		case chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE:
			l.handleCloseUserSession(context.Background(), conn, event.Data, isGuest)
		case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING, chat.ChatEventType_CHAT_EVENT_TYPE_STOP_TYPING:
			l.handleUserTyping(context.Background(), conn, event.Type, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT:
			l.handleSubmitEvaluation(context.Background(), conn, event.Data, isGuest)
		case chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT:
			conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT, map[string]int64{"time": time.Now().UnixMilli()})
		default:
			sendWSError(conn, "unsupported event type")
		}
	}
}

// 处理用户断开连接。
func (l *MessagesLogic) onClose(isGuest bool) func(*ws.Connection) {
	return func(conn *ws.Connection) {
		if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
			return
		}
		l.closeUserSession(context.Background(), conn, isGuest, "用户已离开客服页面")
	}
}

// 处理用户发消息：
// 游客：chat-api 构造消息并转发；
// 登录用户：先调用内部服务入库，再把返回的消息转发给后台和用户侧。
func (l *MessagesLogic) handleSendUserMessage(ctx context.Context, conn *ws.Connection, payload json.RawMessage, isGuest bool) {
	var data UserMessagePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		sendWSError(conn, "invalid message payload")
		return
	}
	if strings.TrimSpace(conn.SessionNo) == "" {
		sendWSError(conn, "sessionNo is required")
		return
	}

	var msg *chat.ChatMessage
	if isGuest {
		msg = buildChatMessage(conn, chat.ChatSenderType_CHAT_SENDER_TYPE_USER, conn.UserId, data)
	} else {
		req := chat.SendUserMessageReq{
			SessionNo:       conn.SessionNo,
			MessageType:     normalizeMessageType(int64(data.MessageType)),
			Content:         strings.TrimSpace(data.Content),
			Url:             data.Url,
			FileName:        data.FileName,
			FileSize:        data.FileSize,
			MimeType:        data.MimeType,
			SenderNickname:  firstNonEmpty(conn.Username, data.SenderNickname),
			SenderAvatarUrl: firstNonEmpty(conn.AvatarUrl, data.SenderAvatarUrl),
		}
		resp, err := l.svcCtx.ChatAppCli.SendUserMessage(contextWithChatIdentity(ctx, conn.MerchantId, conn.UserId), &req)
		if err != nil {
			sendWSError(conn, err.Error())
			return
		}
		msg = resp.GetData()
	}

	if msg == nil {
		sendWSError(conn, "message data is empty")
		return
	}
	normalizeOutgoingMessage(conn, msg, chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE)
	if err := l.publishMessageEvent(ctx, msg); err != nil {
		sendWSError(conn, err.Error())
		return
	}
	l.sendMessageAckEvent(conn, msg)
}

func (l *MessagesLogic) handleUserTyping(ctx context.Context, conn *ws.Connection, eventType chat.ChatEventType, payload json.RawMessage) {
	if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
		return
	}
	message := "用户正在输入"
	if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_STOP_TYPING {
		message = "用户停止输入"
	}
	event := &chat.ChatMessageEvent{
		Type:      eventType,
		CreatedAt: time.Now().UnixMilli(),
		Data:      buildSystemChatMessage(conn, eventType, message),
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  conn.SessionNo,
			MerchantId: conn.MerchantId,
			UserId:     conn.UserId,
			Message:    message,
			CreatedAt:  time.Now().UnixMilli(),
		},
	}
	if len(payload) > 0 {
		var data struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(payload, &data); err == nil && strings.TrimSpace(data.Message) != "" {
			event.SessionEvent.Message = strings.TrimSpace(data.Message)
			event.Data.Content = event.SessionEvent.Message
		}
	}
	if err := l.publishChatEvent(ctx, event); err != nil {
		sendWSError(conn, err.Error())
	}
}

func (l *MessagesLogic) handleSubmitEvaluation(ctx context.Context, conn *ws.Connection, payload json.RawMessage, isGuest bool) {
	if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
		sendWSError(conn, "sessionNo is required")
		return
	}
	var data struct {
		Score   int32  `json:"score"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		sendWSError(conn, "invalid evaluation payload")
		return
	}
	if data.Score < 1 || data.Score > 5 {
		sendWSError(conn, "score must be between 1 and 5")
		return
	}
	if !isGuest {
		resp, err := l.svcCtx.ChatAppCli.SubmitChatSatisfaction(contextWithChatIdentity(ctx, conn.MerchantId, conn.UserId), &chat.SubmitChatSatisfactionReq{
			SessionNo: conn.SessionNo,
			Score:     data.Score,
			Content:   strings.TrimSpace(data.Content),
			Tags:      strings.TrimSpace(data.Tags),
		})
		if err != nil {
			sendWSError(conn, err.Error())
			return
		}
		if resp.GetBase().GetCode() != successCode {
			sendWSError(conn, resp.GetBase().GetMsg())
			return
		}
		conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT, resp)
		return
	}
	msg := buildSystemChatMessage(conn, chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT, "用户已提交评价")
	msg.MessageType = chat.ChatMessageType_CHAT_MESSAGE_TYPE_EVALUATION
	msg.Extra = stringFromMap(map[string]interface{}{
		"score":   data.Score,
		"content": strings.TrimSpace(data.Content),
		"tags":    strings.TrimSpace(data.Tags),
	})
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT,
		CreatedAt: time.Now().UnixMilli(),
		Data:      msg,
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  conn.SessionNo,
			MerchantId: conn.MerchantId,
			UserId:     conn.UserId,
			Message:    "用户已提交评价",
			CreatedAt:  time.Now().UnixMilli(),
		},
	}
	if err := l.publishChatEvent(ctx, event); err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT, map[string]string{"message": "ok"})
}

func (l *MessagesLogic) handleCloseUserSession(ctx context.Context, conn *ws.Connection, payload json.RawMessage, isGuest bool) {
	closeReason := "用户主动结束会话"
	if len(payload) > 0 {
		var data struct {
			CloseReason string `json:"closeReason"`
		}
		if err := json.Unmarshal(payload, &data); err != nil {
			sendWSError(conn, "invalid session close payload")
			return
		}
		closeReason = firstNonEmpty(data.CloseReason, closeReason)
	}
	l.closeUserSession(ctx, conn, isGuest, closeReason)
}

func (l *MessagesLogic) closeUserSession(ctx context.Context, conn *ws.Connection, isGuest bool, reason string) {
	if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
		return
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "用户已离开客服页面"
	}

	// 游客：只转发关闭事件。
	// 登录用户：先更新会话状态，再转发关闭事件给后台和用户侧。
	if !isGuest {
		_, err := l.svcCtx.ChatAppCli.CloseMyChatSession(contextWithChatIdentity(ctx, conn.MerchantId, conn.UserId), &chat.CloseMyChatSessionReq{
			SessionNo:   conn.SessionNo,
			CloseReason: reason,
		})
		if err != nil {
			logx.Errorf("close chat ws persistent session failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
		}
	}

	if err := l.publishSessionCloseEvent(ctx, conn, reason); err != nil {
		logx.Errorf("publish chat session close event failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
	}
}

func sendWSError(conn *ws.Connection, message string) {
	if conn == nil {
		return
	}
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

func buildChatMessage(conn *ws.Connection, senderType chat.ChatSenderType, senderId int64, data UserMessagePayload) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	senderNickname := firstNonEmpty(data.SenderNickname, conn.Username, guestUsername)
	return &chat.ChatMessage{
		MessageNo:   nextGuestNo(guestMessagePrefix),
		SessionNo:   conn.SessionNo,
		EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		MessageType: normalizeMessageType(int64(data.MessageType)),
		Sender: &chat.ChatMessageUser{
			Id:        senderId,
			Type:      senderType,
			Nickname:  senderNickname,
			AvatarUrl: firstNonEmpty(data.SenderAvatarUrl, conn.AvatarUrl),
		},
		ClientMessageId: strings.TrimSpace(data.ClientMessageId),
		Content:         strings.TrimSpace(data.Content),
		Url:             data.Url,
		FileName:        data.FileName,
		FileSize:        data.FileSize,
		MimeType:        data.MimeType,
		Width:           int32(data.Width),
		Height:          int32(data.Height),
		Duration:        int32(data.Duration),
		ReplyMessageId:  strings.TrimSpace(data.ReplyMessageId),
		Extra:           strings.TrimSpace(data.Extra),
		Status:          chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		Self:            false,
		NeedAck:         true,
		CreateTime:      now,
		UpdateTime:      now,
	}
}

func buildSystemChatMessage(conn *ws.Connection, eventType chat.ChatEventType, content string) *chat.ChatMessage {
	now := time.Now().UnixMilli()
	return &chat.ChatMessage{
		MessageNo:   nextGuestNo(guestMessagePrefix),
		SessionNo:   conn.SessionNo,
		EventType:   eventType,
		MessageType: chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT,
		Sender: &chat.ChatMessageUser{
			Type:     chat.ChatSenderType_CHAT_SENDER_TYPE_SYSTEM,
			Nickname: systemNickname,
		},
		Content:    strings.TrimSpace(content),
		Status:     chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT,
		CreateTime: now,
		UpdateTime: now,
	}
}

func (l *MessagesLogic) sendConnectedEvent(conn *ws.Connection) {
	if conn == nil {
		return
	}
	now := time.Now().UnixMilli()
	conn.SendEvent(&chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM,
		CreatedAt: now,
		Data:      buildSystemChatMessage(conn, chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM, "连接建立成功"),
	})
}

func (l *MessagesLogic) sendMessageAckEvent(conn *ws.Connection, msg *chat.ChatMessage) {
	if conn == nil || msg == nil {
		return
	}
	conn.SendEvent(&chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_DELIVERED,
		CreatedAt: time.Now().UnixMilli(),
		Data:      msg,
	})
}

func (l *MessagesLogic) publishMessageEvent(ctx context.Context, msg *chat.ChatMessage) error {
	if msg == nil {
		return fmt.Errorf("message data is empty")
	}
	if msg.EventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN {
		msg.EventType = chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		Data:      msg,
		CreatedAt: time.Now().UnixMilli(),
	}
	return l.publishChatEvent(ctx, event)
}

func (l *MessagesLogic) publishSessionCloseEvent(ctx context.Context, conn *ws.Connection, message string) error {
	now := time.Now().UnixMilli()
	systemMessage := buildSystemChatMessage(conn, chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE, message)
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE,
		CreatedAt: now,
		Data:      systemMessage,
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  conn.SessionNo,
			MerchantId: conn.MerchantId,
			UserId:     conn.UserId,
			Status:     chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED,
			Message:    strings.TrimSpace(message),
			CreatedAt:  now,
		},
	}
	return l.publishChatEvent(ctx, event)
}

func (l *MessagesLogic) sendQueueEvent(conn *ws.Connection, eventType chat.ChatEventType, message string) {
	if conn == nil {
		return
	}
	now := time.Now().UnixMilli()
	event := &chat.ChatMessageEvent{
		Type:      eventType,
		CreatedAt: now,
		Data:      buildSystemChatMessage(conn, eventType, message),
		Queue: &chat.ChatQueueInfo{
			MerchantId:  conn.MerchantId,
			SessionNo:   conn.SessionNo,
			UserId:      conn.UserId,
			Message:     message,
			UpdateTimes: now,
		},
	}
	conn.SendEvent(event)
}

func (l *MessagesLogic) publishUserOnlineEvent(ctx context.Context, conn *ws.Connection, isGuest bool) {
	if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
		logx.Errorf("")
		return
	}

	now := time.Now().UnixMilli()
	session := &chat.ChatSession{
		SessionNo:       conn.SessionNo,
		MerchantId:      conn.MerchantId,
		UserId:          conn.UserId,
		Source:          chat.ChatSessionSource_CHAT_SESSION_SOURCE_WEB,
		Status:          chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING,
		Priority:        chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL,
		Title:           firstNonEmpty(conn.Username, guestUsername),
		LastMessageTime: now,
		CreateTimes:     now,
		UpdateTimes:     now,
		IsGuest:         isGuest,
		AvatarUrl:       conn.AvatarUrl,
	}
	event := &chat.ChatMessageEvent{
		Type:      chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		CreatedAt: now,
		Session:   session,
		SessionEvent: &chat.ChatSessionEvent{
			SessionNo:  conn.SessionNo,
			MerchantId: conn.MerchantId,
			UserId:     conn.UserId,
			Status:     chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING,
			Message:    "用户进入客服，等待接待",
			Session:    session,
			CreatedAt:  now,
		},
	}
	if err := l.publishChatEvent(ctx, event); err != nil {
		logx.Errorf("publish chat user join event failed: %v", err)
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

func normalizeOutgoingMessage(conn *ws.Connection, msg *chat.ChatMessage, eventType chat.ChatEventType) {
	if msg == nil {
		return
	}
	now := time.Now().UnixMilli()
	if msg.MessageNo == "" {
		msg.MessageNo = nextGuestNo(guestMessagePrefix)
	}
	if msg.SessionNo == "" && conn != nil {
		msg.SessionNo = conn.SessionNo
	}
	if msg.EventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNKNOWN {
		msg.EventType = eventType
	}
	if msg.MessageType == chat.ChatMessageType_CHAT_MESSAGE_TYPE_UNKNOWN {
		msg.MessageType = chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT
	}
	if msg.Status == chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_UNKNOWN {
		msg.Status = chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT
	}
	if msg.Sender == nil && conn != nil {
		msg.Sender = &chat.ChatMessageUser{
			Id:        conn.UserId,
			Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
			Nickname:  firstNonEmpty(conn.Username, guestUsername),
			AvatarUrl: conn.AvatarUrl,
		}
	}
	if msg.CreateTime == 0 {
		msg.CreateTime = now
	}
	if msg.UpdateTime == 0 {
		msg.UpdateTime = now
	}
}

type MessageTypeValue int64

func (m *MessageTypeValue) UnmarshalJSON(raw []byte) error {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		*m = MessageTypeValue(chat.ChatMessageType_CHAT_MESSAGE_TYPE_UNKNOWN)
		return nil
	}
	if value[0] == '"' {
		var name string
		if err := json.Unmarshal(raw, &name); err != nil {
			return err
		}
		name = strings.TrimSpace(strings.ToUpper(name))
		name = strings.TrimPrefix(name, "CHAT_MESSAGE_TYPE_")
		fullName := "CHAT_MESSAGE_TYPE_" + name
		if n, ok := chat.ChatMessageType_value[fullName]; ok {
			*m = MessageTypeValue(n)
			return nil
		}
		return fmt.Errorf("unsupported message type: %s", name)
	}
	n, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid message type")
	}
	*m = MessageTypeValue(n)
	return nil
}

type UserMessagePayload struct {
	MessageType     MessageTypeValue `json:"messageType"`
	ClientMessageId string           `json:"clientMessageId"`
	Content         string           `json:"content"`
	Url             string           `json:"url"`
	FileName        string           `json:"fileName"`
	FileSize        int64            `json:"fileSize"`
	MimeType        string           `json:"mimeType"`
	Width           int64            `json:"width"`
	Height          int64            `json:"height"`
	Duration        int64            `json:"duration"`
	ReplyMessageId  string           `json:"replyMessageId"`
	SenderNickname  string           `json:"senderNickname"`
	SenderAvatarUrl string           `json:"senderAvatarUrl"`
	Extra           string           `json:"extra"`
}

func normalizeMessageType(messageType int64) chat.ChatMessageType {
	mt := chat.ChatMessageType(messageType)
	if mt == chat.ChatMessageType_CHAT_MESSAGE_TYPE_UNKNOWN {
		return chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT
	}
	return mt
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}

func stringFromMap(payload map[string]interface{}) string {
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}
