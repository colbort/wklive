package chat_ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"
	"chat-admin-api/internal/ws"
	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

func (l *MessagesLogic) Messages(w http.ResponseWriter, r *http.Request, req types.ChatAdminWSMessagesReq) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("upgrade chat admin ws failed, userId=%d err=%v", req.UserId, err)
		return
	}

	user, err := l.svcCtx.ChatAdminCli.GetChatUserById(l.ctx, &chat.GetChatUserByIdReq{Id: req.UserId})
	if err != nil {
		logx.Errorf("upgrade chat admin ws failed, userId=%d err=%v", req.UserId, err)
		return
	}

	streamCtx, streamCancel := context.WithCancel(l.ctx)
	client := ws.NewConnection(
		conn,
		user.Data.Nickname,
		user.Data.AvatarUrl,
		req.MerchantId,
		req.AgentId,
		req.UserId,
		req.SessionNo,
		l.onMessage(),
		func(*ws.Connection) {
			streamCancel()
		},
	)
	client.SendEvent(&chat.ChatMessageEvent{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM_NOTICE,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatMessageEvent_SystemNotice{
			SystemNotice: &chat.ChatSystemNoticePayload{
				SessionNo:  req.SessionNo,
				Title:      "connection",
				Content:    "chat admin websocket connected",
				Level:      "info",
				ShowInChat: false,
			},
		},
	})

	go l.subscribeStream(streamCtx, client)
	go client.WritePump()
	client.ReadPump()
}

func (l *MessagesLogic) onMessage() func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		eventType, ok := chat.ChatEventType_value[event.EventType]
		if !ok {
			sendWSError(conn, "event type parse err: "+event.EventType)
			return
		}
		switch chat.ChatEventType(eventType) {
		case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ACCEPTED: // 接待客户服务
			l.handleAcceptChatSession(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE: //
			l.handleSendAgentMessage(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM_NOTICE:
			// TODO handle merchant system notice
		case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_LEAVE:
			// TODO handle agent leave
		case chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REQUEST:
			// TODO handle transfer request
		case chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_ACCEPT:
			// TODO handle transfer accept
		case chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REJECT:
			// TODO handle transfer reject
		case chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE:
			l.handleCloseChatSession(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE:
			l.handleEvaluationInvite(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING:
			l.handleAgentTyping(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELIVERED:
			// TODO handle message delivered
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_READ:
			// TODO handle message read
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL:
			// TODO handle message recall
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELETE:
			// TODO handle message delete
		case chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT:
			conn.SendEvent(&chat.ChatMessageEvent{
				Code:      200,
				Msg:       "",
				EventType: chat.ChatEventType_CHAT_EVENT_TYPE_HEARTBEAT,
				CreatedAt: utils.NowMillis(),
				Payload: &chat.ChatMessageEvent_Heartbeat{
					Heartbeat: &chat.ChatHeartbeatPayload{},
				},
			})
		default:
			sendWSError(conn, "unsupported event type")
		}
	}
}

func (l *MessagesLogic) subscribeStream(ctx context.Context, conn *ws.Connection) {
	if conn == nil || l.svcCtx == nil || l.svcCtx.ChatAdminCli == nil {
		return
	}
	stream, err := l.svcCtx.ChatAdminCli.AdminSubscribeStream(ctx, &chat.AdminChatSubscribeRequest{
		MerchantId: conn.MerchantId,
		UserId:     conn.AgentUserId,
		AgentId:    conn.AgentId,
		SessionNo:  conn.SessionNo,
		Admin:      true,
	})
	if err != nil {
		logx.Errorf("subscribe chat admin stream failed, merchantId=%d userId=%d agentId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.AgentId, conn.SessionNo, err)
		return
	}
	for {
		event, err := stream.Recv()
		if err != nil {
			if ctx.Err() == nil {
				logx.Errorf("receive chat admin stream failed, merchantId=%d userId=%d agentId=%d sessionNo=%s err=%v", conn.MerchantId, conn.UserId, conn.AgentId, conn.SessionNo, err)
			}
			return
		}
		if event.GetEventType() == chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN {
			session := event.GetSession()
			if session != nil {
				conn.SessionNo = session.SessionNo
				conn.IsGuest = session.IsGuest
				conn.UserId = session.UserId
			}
		}
		conn.SendEvent(event)
	}
}

func (l *MessagesLogic) handleSendAgentMessage(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	var req chat.SendAgentMessageReq
	if err := json.Unmarshal(payload, &req); err != nil {
		sendWSError(conn, "invalid send_agent_message payload")
		return
	}
	req.Sender = &chat.ChatMessageUser{
		Id:        conn.UserId,
		Nickname:  conn.Nickname,
		AvatarUrl: conn.AvatarUrl,
		Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT,
	}
	req.SessionNo = conn.SessionNo
	req.MerchantId = conn.MerchantId
	req.IsGuest = conn.IsGuest
	_, err := l.svcCtx.ChatAdminCli.SendAgentMessage(ctx, &req)
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
}

func (l *MessagesLogic) handleAcceptChatSession(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	var data chat.AcceptChatSessionReq
	if err := json.Unmarshal(payload, &data); err != nil {
		sendWSError(conn, "invalid accept_chat_session payload")
		return
	}

	resp, err := l.svcCtx.ChatAdminCli.AcceptChatSession(ctx, &chat.AcceptChatSessionReq{
		SessionNo:  conn.SessionNo,
		Reason:     "accept",
		MerchantId: conn.MerchantId,
		AgentId:    conn.AgentId,
		IsGuest:    conn.IsGuest,
	})
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendEvent(&chat.ChatMessageEvent{
		Code:      resp.Base.Code,
		Msg:       resp.Base.Msg,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ACCEPTED,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatMessageEvent_Session{
			Session: resp.Data,
		},
	})
}

func (l *MessagesLogic) handleCloseChatSession(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	var req chat.CloseChatSessionReq
	if err := json.Unmarshal(payload, &req); err != nil {
		sendWSError(conn, "invalid close_chat_session payload")
		return
	}
	req.SessionNo = conn.SessionNo
	req.CloseReason = firstNonEmpty(req.CloseReason, "closed by agent")
	req.IsGuest = conn.IsGuest

	resp, err := l.svcCtx.ChatAdminCli.CloseChatSession(ctx, &req)
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendEvent(&chat.ChatMessageEvent{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatMessageEvent_Session{
			Session: resp.Data,
		},
	})
}

func (l *MessagesLogic) handleAgentTyping(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	conn.SendEvent(&chat.ChatMessageEvent{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_TYPING,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatMessageEvent_Typing{
			Typing: &chat.ChatTypingPayload{},
		},
	})
}

func (l *MessagesLogic) handleEvaluationInvite(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	conn.SendEvent(&chat.ChatMessageEvent{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatMessageEvent_Evaluation{
			Evaluation: &chat.ChatEvaluationPayload{},
		},
	})
}

func sessionAgentFromPayload(conn *ws.Connection, sessionNo string, agentId int64) (string, int64) {
	sessionNo = strings.TrimSpace(sessionNo)
	if sessionNo == "" {
		sessionNo = conn.SessionNo
	}
	if agentId == 0 {
		agentId = conn.AgentId
	}
	return sessionNo, agentId
}

func sendWSError(conn *ws.Connection, message string) {
	conn.SendEvent(&chat.ChatMessageEvent{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_ERROR,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatMessageEvent_Error{
			Error: &chat.ChatErrorPayload{
				SessionNo:    conn.SessionNo,
				MessageNo:    "",
				ErrorCode:    0,
				ErrorMessage: message,
				Detail:       message,
				Retryable:    false,
			},
		},
	})
}

func (l *MessagesLogic) appendTransientMessage(ctx context.Context, merchantId int64, eventType chat.ChatEventType, msg *chat.ChatMessage, session *chat.ChatSession) error {
	if l.svcCtx == nil || l.svcCtx.ChatAdminCli == nil {
		return nil
	}
	resp, err := l.svcCtx.ChatAdminCli.SendAgentMessage(ctx, &chat.SendAgentMessageReq{
		MerchantId: merchantId,
		EventType:  eventType,
		Message:    msg,
		Session:    session,
		IsGuest:    true,
	})
	if err != nil {
		return err
	}
	if resp.GetBase().GetCode() != 200 {
		return errors.New(resp.GetBase().GetMsg())
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}
