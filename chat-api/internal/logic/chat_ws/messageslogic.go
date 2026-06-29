package chat_ws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"chat-api/internal/svc"
	"chat-api/internal/types"
	"chat-api/internal/ws"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
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
	ext := map[string]string{
		"nickname":  req.Nickname,
		"avatarUrl": req.AvatarUrl,
	}
	extJson, err := json.Marshal(ext)
	if err != nil {
		logx.Errorf("json marshal err, merchantId=%d userId=%d guest=%t", req.MerchantId, req.UserId, req.IsGuest)
		_ = conn.Close()
		return
	}
	resp, err := l.svcCtx.ChatAppCli.OpenChatSession(l.ctx, &chat.OpenChatSessionReq{
		Source:     chat.ChatSessionSource_CHAT_SESSION_SOURCE_WEB,
		MerchantId: req.MerchantId,
		UserId:     req.UserId,
		IsGuest:    req.IsGuest,
		SessionNo:  req.SessionNo,
		ExtJson:    string(extJson),
	})
	if err != nil {
		logx.Errorf("open chat ws session failed, merchantId=%d userId=%d guest=%t err=%v", req.MerchantId, req.UserId, req.IsGuest, err)
		_ = conn.Close()
		return
	}
	if resp.GetBase().GetCode() != successCode {
		logx.Errorf("open chat ws session rejected, merchantId=%d userId=%d guest=%t code=%d msg=%s", req.MerchantId, req.UserId, req.IsGuest, resp.GetBase().GetCode(), resp.GetBase().GetMsg())
		_ = conn.Close()
		return
	}
	sessionNo := resp.Data.SessionNo
	if sessionNo == "" {
		logx.Errorf("session is empty, merchantId=%d userId=%d guest=%t", req.MerchantId, req.UserId, req.IsGuest)
		_ = conn.Close()
		return
	}

	streamCtx, streamCancel := context.WithCancel(l.ctx)
	client := ws.NewConnection(
		conn,
		req.UserId,
		req.Nickname,
		req.AvatarUrl,
		req.MerchantId,
		sessionNo,
		l.onMessage(req.IsGuest),
		func(conn *ws.Connection) {
			streamCancel()
			l.onClose(req.IsGuest)(conn)
		},
	)

	// 读 RPC 消息
	go l.subscribeStream(streamCtx, client, req.IsGuest)
	// 读客户端消息
	go client.WritePump()
	// 先客户端写消息
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
		case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING:
			l.handleUserTyping(context.Background(), conn, event.Data, isGuest)
		case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT:
			l.handleSubmitEvaluation(context.Background(), conn, event.Data, isGuest)
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_READ:
			// TODO handle message read
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL:
			// TODO handle message recall
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELETE:
			// TODO handle message delete
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
		_, err := l.svcCtx.ChatAppCli.CloseMyChatSession(context.Background(), &chat.CloseMyChatSessionReq{
			SessionNo:   conn.SessionNo,
			MerchantId:  conn.MerchantId,
			UserId:      conn.UserId,
			CloseReason: "用户已离开客服页面",
			IsGuest:     isGuest,
		})
		if err != nil {
			logx.Errorf("close chat ws persistent session failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
		}
	}
}

func (l *MessagesLogic) subscribeStream(ctx context.Context, conn *ws.Connection, isGuest bool) {
	if conn == nil || l.svcCtx == nil || l.svcCtx.ChatAppCli == nil {
		logx.Error("app subscribe err")
		return
	}
	stream, err := l.svcCtx.ChatAppCli.AppSubscribeStream(ctx, &chat.AppChatSubscribeRequest{
		MerchantId: conn.MerchantId,
		UserId:     conn.UserId,
		SessionNo:  conn.SessionNo,
		IsGuest:    isGuest,
	})
	fmt.Println("============================= 77")
	if err != nil {
		logx.Errorf("subscribe chat app stream failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
		return
	}
	fmt.Println("============================= 88")
	for {
		event, err := stream.Recv()
		if err != nil {
			if ctx.Err() == nil {
				logx.Errorf("receive chat app stream failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
			}
			return
		}
		fmt.Println("============================= 99")
		conn.SendEvent(event)
	}
}

// 处理用户发消息：
// 游客：chat-api 构造消息并转发；
// 登录用户：先调用内部服务入库，再把返回的消息转发给后台和用户侧。
func (l *MessagesLogic) handleSendUserMessage(ctx context.Context, conn *ws.Connection, payload json.RawMessage, isGuest bool) {
	var req chat.SendUserMessageReq
	if err := json.Unmarshal(payload, &req); err != nil {
		sendWSError(conn, "invalid message payload")
		return
	}
	if strings.TrimSpace(conn.SessionNo) == "" {
		sendWSError(conn, "sessionNo is required")
		return
	}
	req.Sender = &chat.ChatMessageUser{
		Id:        conn.UserId,
		Nickname:  conn.Username,
		AvatarUrl: conn.AvatarUrl,
		Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
	}
	req.MerchantId = conn.MerchantId
	req.IsGuest = isGuest
	resp, err := l.svcCtx.ChatAppCli.SendUserMessage(ctx, &req)
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	if resp.GetBase().GetCode() != successCode {
		sendWSError(conn, resp.GetBase().GetMsg())
		return
	}
	msg := resp.GetData()
	if msg == nil {
		sendWSError(conn, "message data is empty")
		return
	}

	now := time.Now().UnixMilli()
	conn.SendEvent(&chat.ChatMessageEvent{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELIVERED,
		CreatedAt: now,
		Payload: &chat.ChatMessageEvent_Receipt{Receipt: &chat.ChatMessageReceiptPayload{
			SessionNo:     conn.SessionNo,
			MessageNo:     msg.MessageNo,
			SenderId:      req.Sender.Id,
			OperatorId:    conn.UserId,
			OperatorType:  chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
			MessageStatus: chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_DELIVERED,
			ReceiptTime:   now,
		}},
	})
}

func (l *MessagesLogic) handleUserTyping(ctx context.Context, conn *ws.Connection, payload json.RawMessage, isGuest bool) {
	if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
		return
	}
	now := time.Now().UnixMilli()
	typing := chat.ChatTypingPayload{
		SessionNo:  conn.SessionNo,
		SenderId:   strconv.FormatInt(conn.UserId, 10),
		SenderType: chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
		Text:       "用户正在输入",
		ActionTime: now,
	}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &typing); err != nil {
			sendWSError(conn, "invalid typing payload")
			return
		}
	}
	resp, err := l.svcCtx.ChatAppCli.SendUserTyping(ctx, &chat.SendUserTypingReq{
		MerchantId: conn.MerchantId,
		UserId:     conn.UserId,
		IsGuest:    isGuest,
		Typing:     &typing,
	})
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	if resp.GetBase().GetCode() != successCode {
		sendWSError(conn, resp.GetBase().GetMsg())
		return
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
	resp, err := l.svcCtx.ChatAppCli.SubmitChatSatisfaction(ctx, &chat.SubmitChatSatisfactionReq{
		SessionNo:  conn.SessionNo,
		MerchantId: conn.MerchantId,
		UserId:     conn.UserId,
		Score:      data.Score,
		Content:    strings.TrimSpace(data.Content),
		Tags:       strings.TrimSpace(data.Tags),
		IsGuest:    isGuest,
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
}

func (l *MessagesLogic) handleCloseUserSession(ctx context.Context, conn *ws.Connection, payload json.RawMessage, isGuest bool) {
	_, err := l.svcCtx.ChatAppCli.CloseMyChatSession(context.Background(), &chat.CloseMyChatSessionReq{
		SessionNo:   conn.SessionNo,
		MerchantId:  conn.MerchantId,
		UserId:      conn.UserId,
		CloseReason: "用户主动结束会话",
		IsGuest:     isGuest,
	})
	if err != nil {
		logx.Errorf("close chat ws persistent session failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
	}
}

func sendWSError(conn *ws.Connection, message string) {
	if conn == nil {
		return
	}
	conn.SendEvent(&chat.ChatMessageEvent{
		Code:      1,
		Msg:       message,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_ERROR,
		CreatedAt: time.Now().UnixMilli(),
		Payload: &chat.ChatMessageEvent_Error{Error: &chat.ChatErrorPayload{
			SessionNo:    conn.SessionNo,
			ErrorCode:    1,
			ErrorMessage: message,
			Retryable:    false,
		}},
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}
