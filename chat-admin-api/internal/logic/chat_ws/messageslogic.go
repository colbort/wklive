package chat_ws

import (
	"context"
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
		&chat.ChatMessageUser{
			Id:        req.UserId,
			Nickname:  user.Data.Nickname,
			AvatarUrl: user.Data.AvatarUrl,
		},
		req.MerchantId,
		req.AgentId,
		l.onMessage(),
		func(*ws.Connection) {
			streamCancel()
		},
	)
	client.SendEvent(&chat.ChatWsResponse{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_WS_CONNECTED,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatWsResponse_Connected{
			Connected: &chat.WsConnectedPayload{
				Message:    "chat admin websocket connected",
				SessionNo:  "",
				MerchantId: req.MerchantId,
				UserId:     req.UserId,
				Nickname:   user.Data.Nickname,
				AvatarUrl:  user.Data.AvatarUrl,
			},
		},
	})

	go l.subscribeStream(streamCtx, client)
	go client.WritePump()
	client.ReadPump()
}

func (l *MessagesLogic) onMessage() func(*ws.Connection, *chat.ChatWsRequest) {
	return func(conn *ws.Connection, event *chat.ChatWsRequest) {
		if event == nil {
			conn.SendError("invalid request", "invalid request")
			return
		}
		eventType := event.GetEventType()
		if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_UNSPECIFIED {
			conn.SendError("event type is required", "event type is required")
			return
		}
		switch eventType {
		case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ACCEPTED: // 接待客户服务
			l.handleAcceptChatSession(context.Background(), conn, event.GetAgent())
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE: //
			l.handleSendAgentMessage(context.Background(), conn, event.GetMessage())
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
			l.handleCloseChatSession(context.Background(), conn, event.GetSession())
		case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE:
			l.handleEvaluationInvite(context.Background(), conn, event.GetEvaluation())
		case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING:
			l.handleAgentTyping(context.Background(), conn, event.GetTyping())
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELIVERED:
			// TODO handle message delivered
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
				Payload: &chat.ChatWsResponse_Heartbeat{
					Heartbeat: &chat.ChatHeartbeatPayload{},
				},
			})
		default:
			conn.SendError("unsupported event type", "unsupported event type: "+event.EventType.String())
		}
	}
}

func (l *MessagesLogic) subscribeStream(ctx context.Context, conn *ws.Connection) {
	if conn == nil || l.svcCtx == nil || l.svcCtx.ChatAdminCli == nil {
		return
	}
	stream, err := l.svcCtx.ChatAdminCli.AdminSubscribeStream(ctx, &chat.AdminChatSubscribeRequest{})
	if err != nil {
		logx.Errorf("subscribe chat admin stream failed, merchantId=%d agentId=%d err=%v", conn.MerchantId, conn.AgentId, err)
		return
	}
	for {
		event, err := stream.Recv()
		if err != nil {
			if ctx.Err() == nil {
				logx.Errorf("receive chat admin stream failed, merchantId=%d agentId=%d err=%v", conn.MerchantId, conn.AgentId, err)
			}
			return
		}
		if event.EventType == chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE {
			userState := event.GetUserState()
			if userState != nil {
				conn.Receivers.Delete(userState.SessionNo)
			}
			logx.Infof("user %d finish session, session no is %s", userState.UserId, userState.SessionNo)
		}

		conn.SendEvent(event)
	}
}

func (l *MessagesLogic) handleSendAgentMessage(ctx context.Context, conn *ws.Connection, payload *chat.ChatWsMessagePayload) {
	if payload == nil {
		conn.SendError("invalid send_agent_message payload", "invalid send_agent_message payload")
		return
	}
	req := chat.SendAgentMessageReq{
		SessionNo:       payload.GetSessionNo(),
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
		Sender:          payload.Sender,
		MerchantId:      conn.MerchantId,
		IsGuest:         conn.IsGuest,
	}
	receiver, ok := conn.Receivers.Load(payload.SessionNo)
	if ok {
		req.Receiver = receiver.(*chat.ChatMessageUser)
	}
	resp, err := l.svcCtx.ChatAdminCli.SendAgentMessage(ctx, &req)
	if err != nil {
		conn.SendError("failed to send agent message", err.Error())
		return
	}
	if resp.Base.Code != 200 {
		logx.Errorf("chat admin send message fail, error message is : %s", resp.Base.Msg)
	} else {
		logx.Errorf("chat admin send message success, message no is : %s", resp.Data.MessageNo)
	}
}

func (l *MessagesLogic) handleAcceptChatSession(ctx context.Context, conn *ws.Connection, payload *chat.ChatWsAgentPayload) {
	if payload == nil {
		conn.SendError("invalid accept_chat_session payload", "invalid accept_chat_session payload")
		return
	}

	resp, err := l.svcCtx.ChatAdminCli.AcceptChatSession(ctx, &chat.AcceptChatSessionReq{
		SessionNo:  payload.GetSessionNo(),
		Reason:     payload.GetReason(),
		MerchantId: conn.MerchantId,
		AgentId:    firstNonZero(payload.GetAgentId(), conn.AgentId),
		IsGuest:    conn.IsGuest,
	})
	if err != nil {
		conn.SendError("failed to accept chat session", err.Error())
		return
	}
	// 加入接待用户
	if resp.Data != nil {
		conn.Receivers.Store(resp.Data.SessionNo, resp.Data.User)
		logx.Infof("user %d accepted, session no is %s\n", resp.Data.User.Id, resp.Data.SessionNo)
	}
	conn.SendEvent(&chat.ChatWsResponse{
		Code:      resp.Base.Code,
		Msg:       resp.Base.Msg,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ACCEPTED,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatWsResponse_Agent{
			Agent: resp.Data.Agent,
		},
	})
}

func (l *MessagesLogic) handleCloseChatSession(ctx context.Context, conn *ws.Connection, payload *chat.ChatWsSessionPayload) {
	if payload == nil {
		conn.SendError("invalid close_chat_session payload", "invalid close_chat_session payload")
		return
	}
	req := chat.CloseChatSessionReq{
		SessionNo:   payload.GetSessionNo(),
		MerchantId:  payload.GetMerchantId(),
		UserId:      payload.GetUserId(),
		CloseReason: payload.GetCloseReason(),
		IsGuest:     payload.GetIsGuest(),
	}
	req.CloseReasonType = chat.ChatSessionCloseReason_CHAT_SESSION_CLOSE_REASON_AGENT
	req.CloseReason = firstNonEmpty(req.CloseReason, "closed by agent")

	resp, err := l.svcCtx.ChatAdminCli.CloseChatSession(ctx, &req)
	if err != nil {
		conn.SendError("failed to close chat session", err.Error())
		return
	}
	conn.SendEvent(&chat.ChatWsResponse{
		Code:      resp.Base.Code,
		Msg:       resp.Base.Msg,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatWsResponse_Session{
			Session: resp.Data,
		},
	})
}

func (l *MessagesLogic) handleAgentTyping(ctx context.Context, conn *ws.Connection, payload *chat.ChatTypingPayload) {
	conn.SendEvent(&chat.ChatWsResponse{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_TYPING,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatWsResponse_Typing{
			Typing: &chat.ChatTypingPayload{},
		},
	})
}

func (l *MessagesLogic) handleEvaluationInvite(ctx context.Context, conn *ws.Connection, payload *chat.ChatEvaluationPayload) {
	conn.SendEvent(&chat.ChatWsResponse{
		Code:      200,
		Msg:       "",
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE,
		CreatedAt: utils.NowMillis(),
		Payload: &chat.ChatWsResponse_Evaluation{
			Evaluation: &chat.ChatEvaluationPayload{},
		},
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

func firstNonZero(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
