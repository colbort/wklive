package chat_ws

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"chat-api/internal/svc"
	"chat-api/internal/types"
	"chat-api/internal/ws"
	"wklive/common/utils"
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
	sessionNo := resp.GetData().GetSessionNo()
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
		req.IsGuest,
		l.onMessage(),
		func(conn *ws.Connection) {
			streamCancel()
			l.onClose()(conn)
		},
	)

	client.SendEvent(&chat.ChatWsResponse{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_WS_CONNECTED,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatWsResponse_Connected{
			Connected: &chat.WsConnectedPayload{
				Message:    "chat app websocket connected",
				SessionNo:  req.SessionNo,
				MerchantId: req.MerchantId,
				UserId:     req.UserId,
				Nickname:   req.Nickname,
				AvatarUrl:  req.AvatarUrl,
				IsGuest:    req.IsGuest,
			},
		},
	})

	client.SendEvent(&chat.ChatWsResponse{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_UPDATE,
		CreatedAt: utils.NowMillis(),
		Payload:   &chat.ChatWsResponse_Queue{Queue: resp.Data},
	})

	// 读 RPC 消息
	go l.subscribeStream(streamCtx, client)
	// 读客户端消息
	go client.WritePump()
	// 先客户端写消息
	client.ReadPump()
}

// 处理用户通过 ws 发送的事件。
func (l *MessagesLogic) onMessage() func(*ws.Connection, *chat.ChatWsRequest) {
	return func(conn *ws.Connection, event *chat.ChatWsRequest) {
		if event == nil {
			conn.SendError("invalid request", "event is nil")
			return
		}
		eventType := event.GetEventType()
		if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNSPECIFIED {
			conn.SendError("event type is required", "event type unsupported")
			return
		}
		switch eventType {
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE:
			l.handleSendUserMessage(context.Background(), conn, event.GetMessage())
		case chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE:
			l.handleCloseUserSession(context.Background(), conn, event.GetSession())
		case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING:
			l.handleUserTyping(context.Background(), conn, event.GetTyping())
		case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT:
			l.handleSubmitEvaluation(context.Background(), conn, event.GetEvaluation())
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_READ:
			// TODO handle message read
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL:
			// TODO handle message recall
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELETE:
			// TODO handle message delete
		case chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT:
			conn.SendEvent(&chat.ChatWsResponse{
				Code:      200,
				Msg:       "",
				EventType: chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT,
				CreatedAt: utils.NowMillis(),
				Payload:   &chat.ChatWsResponse_Agent{},
			})
		default:
			conn.SendError("unsupported event type", "unsupported event type: "+eventType.String())
		}
	}
}

// 处理用户断开连接。
func (l *MessagesLogic) onClose() func(*ws.Connection) {
	return func(conn *ws.Connection) {
		if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
			return
		}
		_, err := l.svcCtx.ChatAppCli.CloseMyChatSession(context.Background(), &chat.CloseMyChatSessionReq{
			SessionNo:       conn.SessionNo,
			MerchantId:      conn.MerchantId,
			UserId:          conn.UserId,
			CloseReasonType: chat.ChatSessionCloseReason_CHAT_SESSION_CLOSE_REASON_INTERNET_ERROR,
			CloseReason:     "用户网络异常断开",
			IsGuest:         conn.IsGuest,
		})
		if err != nil {
			logx.Errorf("close chat ws persistent session failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
		}
	}
}

func (l *MessagesLogic) subscribeStream(ctx context.Context, conn *ws.Connection) {
	if conn == nil || l.svcCtx == nil || l.svcCtx.ChatAppCli == nil {
		logx.Error("app subscribe err")
		return
	}
	stream, err := l.svcCtx.ChatAppCli.AppSubscribeStream(ctx, &chat.AppChatSubscribeRequest{})
	if err != nil {
		logx.Errorf("subscribe chat app stream failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
		return
	}
	for {
		event, err := stream.Recv()
		if err != nil {
			if ctx.Err() == nil {
				logx.Errorf("receive chat app stream failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
			}
			return
		}
		conn.SendEvent(event)
	}
}

// 处理用户发消息：
// 游客：chat-api 构造消息并转发；
// 登录用户：先调用内部服务入库，再把返回的消息转发给后台和用户侧。
func (l *MessagesLogic) handleSendUserMessage(ctx context.Context, conn *ws.Connection, payload *chat.ChatWsMessagePayload) {
	if payload == nil {
		conn.SendError("invalid message payload", "message is nil")
		return
	}
	if strings.TrimSpace(conn.SessionNo) == "" {
		conn.SendError("session no is required", "session no is empty")
		return
	}
	req := chat.SendUserMessageReq{
		SessionNo:       firstNonEmpty(payload.GetSessionNo(), conn.SessionNo),
		ClientMessageId: payload.GetClientMessageId(),
		MessageType:     payload.GetMessageType(),
		Content:         payload.GetContent(),
		Url:             payload.GetUrl(),
		FileName:        payload.GetFileName(),
		FileSize:        payload.GetFileSize(),
		MimeType:        payload.GetMimeType(),
		Width:           payload.GetWidth(),
		Height:          payload.GetHeight(),
		Duration:        payload.GetDuration(),
		Extra:           payload.GetExtra(),
		Sender: &chat.ChatMessageUser{
			Id:        conn.UserId,
			Nickname:  conn.Username,
			AvatarUrl: conn.AvatarUrl,
			Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
		},
		Receiver:   payload.GetReceiver(),
		MerchantId: conn.MerchantId,
		IsGuest:    conn.IsGuest,
	}
	resp, err := l.svcCtx.ChatAppCli.SendUserMessage(ctx, &req)
	if err != nil {
		conn.SendError("send message err", err.Error())
		return
	}
	if resp.GetBase().GetCode() != successCode {
		conn.SendError("send message err", resp.GetBase().GetMsg())
		return
	}
	msg := resp.GetData()
	if msg == nil {
		conn.SendError("message data is empty", "rpc not return message")
		return
	}

	now := time.Now().UnixMilli()
	conn.SendEvent(&chat.ChatWsResponse{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELIVERED,
		CreatedAt: now,
		Payload: &chat.ChatWsResponse_Receipt{Receipt: &chat.ChatMessageReceiptPayload{
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

func (l *MessagesLogic) handleUserTyping(ctx context.Context, conn *ws.Connection, payload *chat.ChatTypingPayload) {
	if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
		return
	}
	now := time.Now().UnixMilli()
	typing := &chat.ChatTypingPayload{
		SessionNo:  conn.SessionNo,
		SenderId:   conn.UserId,
		SenderType: chat.ChatSenderType_CHAT_SENDER_TYPE_USER,
		Text:       "用户正在输入",
		ActionTime: now,
	}
	if payload != nil {
		typing = payload
		typing.SessionNo = firstNonEmpty(typing.GetSessionNo(), conn.SessionNo)
		typing.SenderId = conn.UserId
		typing.SenderType = chat.ChatSenderType_CHAT_SENDER_TYPE_USER
		if typing.ActionTime == 0 {
			typing.ActionTime = now
		}
	}
	resp, err := l.svcCtx.ChatAppCli.SendUserTyping(ctx, &chat.SendUserTypingReq{
		MerchantId: conn.MerchantId,
		UserId:     conn.UserId,
		IsGuest:    conn.IsGuest,
		Typing:     typing,
	})
	if err != nil {
		conn.SendError("call rpc err", err.Error())
		return
	}
	if resp.GetBase().GetCode() != successCode {
		conn.SendError("call rpc fail", resp.GetBase().GetMsg())
		return
	}
}

func (l *MessagesLogic) handleSubmitEvaluation(ctx context.Context, conn *ws.Connection, payload *chat.ChatEvaluationPayload) {
	if conn == nil || strings.TrimSpace(conn.SessionNo) == "" {
		conn.SendError("session no is empty", "session no is required")
		return
	}
	if payload.GetRating() < 1 || payload.GetRating() > 5 {
		conn.SendError("score err", "score must be between 1 and 5")
		return
	}
	resp, err := l.svcCtx.ChatAppCli.SubmitChatSatisfaction(ctx, &chat.SubmitChatSatisfactionReq{
		SessionNo:  conn.SessionNo,
		MerchantId: conn.MerchantId,
		UserId:     conn.UserId,
		Score:      payload.GetRating(),
		Content:    strings.TrimSpace(payload.GetComment()),
		Tags:       strings.Join(payload.GetTags(), ","),
		IsGuest:    conn.IsGuest,
	})
	if err != nil {
		conn.SendError("call rpc fail", err.Error())
		return
	}
	if resp.GetBase().GetCode() != successCode {
		conn.SendError("submit satisfaction err", resp.GetBase().GetMsg())
		return
	}
	conn.SendEvent(&chat.ChatWsResponse{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatWsResponse_Evaluation{
			Evaluation: &chat.ChatEvaluationPayload{},
		},
	})
}

func (l *MessagesLogic) handleCloseUserSession(ctx context.Context, conn *ws.Connection, payload *chat.ChatWsSessionPayload) {
	closeReason := "用户主动结束会话"
	if payload != nil {
		closeReason = firstNonEmpty(payload.GetCloseReason(), closeReason)
	}
	_, err := l.svcCtx.ChatAppCli.CloseMyChatSession(ctx, &chat.CloseMyChatSessionReq{
		SessionNo:       conn.SessionNo,
		CloseReasonType: chat.ChatSessionCloseReason_CHAT_SESSION_CLOSE_REASON_USER,
		CloseReason:     closeReason,
		MerchantId:      conn.MerchantId,
		UserId:          conn.UserId,
		IsGuest:         conn.IsGuest,
	})
	if err != nil {
		logx.Errorf("close chat ws persistent session failed, merchantId=%d userId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.SessionNo, err)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if text := strings.TrimSpace(value); text != "" {
			return text
		}
	}
	return ""
}
