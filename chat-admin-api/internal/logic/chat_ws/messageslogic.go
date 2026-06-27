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
	"wklive/common/helper"
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
		req.UserId,
		user.Data.Nickname,
		user.Data.AvatarUrl,
		req.MerchantId,
		req.AgentId,
		req.SessionNo,
		l.onMessage(),
		func(*ws.Connection) {
			streamCancel()
		},
	)
	client.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM, map[string]interface{}{
		"message":    "chat admin websocket connected",
		"merchantId": req.MerchantId,
		"agentId":    req.AgentId,
		"sessionNo":  req.SessionNo,
	})

	go l.subscribeStream(streamCtx, client)
	go client.WritePump()
	client.ReadPump()
}

func (l *MessagesLogic) onMessage() func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED:
			l.handleAcceptChatSession(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING, chat.ChatEventType_CHAT_EVENT_TYPE_STOP_TYPING:
			l.handleAgentTyping(context.Background(), conn, event.Type, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REQUEST:
		// TODO 会话转接请求
		case chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_ACCEPT:
		// TODO 会话转接接受
		case chat.ChatEventType_CHAT_EVENT_TYPE_TRANSFER_REJECT:
		// TODO 会话转接拒绝
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE:
			l.handleSendAgentMessage(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE:
			l.handleCloseChatSession(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE:
			l.handleEvaluationInvite(context.Background(), conn, event.Data)
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
		UserId:     conn.UserId,
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
		if event.GetEventType() == chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN && event.GetSessionEvent().GetSession() != nil {
			conn.SetChatSession(event.GetSessionEvent().GetSession())
		}
		conn.SendEvent(event)
	}
}

func (l *MessagesLogic) handleSendAgentMessage(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	var data sendAgentMessagePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		sendWSError(conn, "invalid send_agent_message payload")
		return
	}
	applyAgentMessageDefaults(conn, &data)
	req := chat.SendAgentMessageReq{
		SessionNo:   data.SessionNo,
		MessageType: chat.ChatMessageType(data.MessageType),
		Content:     data.Content,
		Url:         data.Url,
		FileName:    data.FileName,
		MimeType:    data.MimeType,
		FileSize:    data.FileSize,
		Width:       int32(data.Width),
		Height:      int32(data.Height),
		Duration:    data.Duration,
		Extra:       data.Extra,
		Sender: &chat.ChatMessageUser{
			Id:        conn.UserId,
			Nickname:  conn.Nickname,
			AvatarUrl: conn.AvatarUrl,
			Type:      chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT,
		},
		MerchantId: conn.MerchantId,
	}
	if conn.IsGuestSession() {
		data.UserId = conn.ChatSessionUserId(data.UserId)
		msg := newTransientAgentMessage(conn.MerchantId, req.SessionNo, conn.UserId, conn.AgentId, conn.Nickname, data)
		if err := l.appendTransientMessage(ctx, conn.MerchantId, chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE, msg, conn.ChatSession()); err != nil {
			sendWSError(conn, err.Error())
			return
		}
		conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	} else {
		resp, err := l.svcCtx.ChatAdminCli.SendAgentMessage(ctx, &req)
		if err != nil {
			sendWSError(conn, err.Error())
			return
		}
		conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE, resp)
	}
}

func (l *MessagesLogic) handleAcceptChatSession(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	var data acceptChatSessionPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		sendWSError(conn, "invalid accept_chat_session payload")
		return
	}
	sessionNo, agentId := sessionAgentFromPayload(conn, data.SessionNo, data.AgentId)
	if sessionNo == "" || agentId == 0 {
		sendWSError(conn, "sessionNo and agentId are required")
		return
	}
	if conn.IsGuestSession() {
		data.UserId = conn.ChatSessionUserId(data.UserId)
		transientSession := conn.ChatSession()
		transientSession.AgentId = agentId
		transientSession.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING
		msg := newTransientSystemMessage(conn.MerchantId, sessionNo, data.UserId, agentId, conn.Nickname+"为您服务")
		if err := l.appendTransientMessage(ctx, conn.MerchantId, chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED, msg, transientSession); err != nil {
			sendWSError(conn, err.Error())
			return
		}
		conn.SetChatSession(transientSession)
		conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	}

	resp, err := l.svcCtx.ChatAdminCli.AcceptChatSession(ctx, &chat.AcceptChatSessionReq{
		SessionNo:  sessionNo,
		Reason:     "accept",
		MerchantId: conn.MerchantId,
		AgentId:    conn.AgentId,
	})
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED, resp)
}

func (l *MessagesLogic) handleCloseChatSession(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	var data closeChatSessionPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		sendWSError(conn, "invalid close_chat_session payload")
		return
	}
	sessionNo := strings.TrimSpace(data.SessionNo)
	if sessionNo == "" {
		sessionNo = conn.SessionNo
	}
	if sessionNo == "" {
		sendWSError(conn, "sessionNo is required")
		return
	}
	reason := firstNonEmpty(data.CloseReason, "closed by agent")
	if conn.IsGuestSession() {
		data.UserId = conn.ChatSessionUserId(data.UserId)
		if err := l.deleteTransientSession(ctx, conn.MerchantId, sessionNo, chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE, "本次会话已结束", data.UserId, conn.AgentId); err != nil {
			sendWSError(conn, err.Error())
			return
		}
		msg := newTransientSystemMessage(conn.MerchantId, sessionNo, data.UserId, conn.AgentId, "本次会话已结束")
		conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	}

	resp, err := l.svcCtx.ChatAdminCli.CloseChatSession(ctx, &chat.CloseChatSessionReq{
		SessionNo:   sessionNo,
		CloseReason: reason,
	})
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE, resp)
}

func (l *MessagesLogic) handleAgentTyping(ctx context.Context, conn *ws.Connection, eventType chat.ChatEventType, payload json.RawMessage) {
	sessionNo := conn.SessionNo
	userId := int64(0)
	message := "客服正在输入"
	if eventType == chat.ChatEventType_CHAT_EVENT_TYPE_STOP_TYPING {
		message = "客服停止输入"
	}
	if len(payload) > 0 {
		var data struct {
			SessionNo string `json:"sessionNo"`
			UserId    int64  `json:"userId"`
			Message   string `json:"message"`
		}
		if err := json.Unmarshal(payload, &data); err == nil {
			if strings.TrimSpace(data.SessionNo) != "" {
				sessionNo = strings.TrimSpace(data.SessionNo)
			}
			userId = data.UserId
			if strings.TrimSpace(data.Message) != "" {
				message = strings.TrimSpace(data.Message)
			}
		}
	}
	if strings.TrimSpace(sessionNo) == "" {
		sendWSError(conn, "sessionNo is required")
		return
	}
	if userId == 0 {
		userId = conn.ChatSessionUserId(0)
	}
	msg := newTransientSystemMessageWithType(conn.MerchantId, sessionNo, userId, conn.AgentId, eventType, chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT, message, "")
	if err := l.publishOnlyTransientMessage(ctx, conn.MerchantId, eventType, msg); err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(eventType, map[string]string{"message": "ok"})
}

func (l *MessagesLogic) handleEvaluationInvite(ctx context.Context, conn *ws.Connection, payload json.RawMessage) {
	sessionNo := conn.SessionNo
	userId := int64(0)
	content := "请对本次服务进行评价"
	if len(payload) > 0 {
		var data struct {
			SessionNo string `json:"sessionNo"`
			UserId    int64  `json:"userId"`
			Content   string `json:"content"`
		}
		if err := json.Unmarshal(payload, &data); err == nil {
			if strings.TrimSpace(data.SessionNo) != "" {
				sessionNo = strings.TrimSpace(data.SessionNo)
			}
			userId = data.UserId
			if strings.TrimSpace(data.Content) != "" {
				content = strings.TrimSpace(data.Content)
			}
		}
	}
	if strings.TrimSpace(sessionNo) == "" {
		sendWSError(conn, "sessionNo is required")
		return
	}
	if userId == 0 {
		userId = conn.ChatSessionUserId(0)
	}
	msg := newTransientSystemMessageWithType(conn.MerchantId, sessionNo, userId, conn.AgentId, chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE, chat.ChatMessageType_CHAT_MESSAGE_TYPE_EVALUATION, content, transientExtra(map[string]interface{}{"action": "invite"}))
	if err := l.appendTransientMessage(ctx, conn.MerchantId, chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE, msg, nil); err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE, map[string]string{"message": "ok"})
}

func applyAgentMessageDefaults(conn *ws.Connection, data *sendAgentMessagePayload) {
	if data.AgentId == 0 {
		data.AgentId = conn.AgentId
	}
	if strings.TrimSpace(data.SessionNo) == "" {
		data.SessionNo = conn.SessionNo
	}
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
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_ERROR, map[string]string{"message": message})
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

func (l *MessagesLogic) publishOnlyTransientMessage(ctx context.Context, merchantId int64, eventType chat.ChatEventType, msg *chat.ChatMessage) error {
	if l.svcCtx == nil || l.svcCtx.ChatAdminCli == nil {
		return nil
	}
	resp, err := l.svcCtx.ChatAdminCli.SendAgentMessage(ctx, &chat.SendAgentMessageReq{
		MerchantId:  merchantId,
		EventType:   eventType,
		Message:     msg,
		PublishOnly: true,
		IsGuest:     true,
	})
	if err != nil {
		return err
	}
	if resp.GetBase().GetCode() != 200 {
		return errors.New(resp.GetBase().GetMsg())
	}
	return nil
}

func (l *MessagesLogic) deleteTransientSession(ctx context.Context, merchantId int64, sessionNo string, eventType chat.ChatEventType, eventMessage string, userId, agentId int64) error {
	if l.svcCtx == nil || l.svcCtx.ChatAdminCli == nil {
		return nil
	}
	resp, err := l.svcCtx.ChatAdminCli.AdminDeleteTransientChatSession(ctx, &chat.AdminDeleteTransientChatSessionReq{
		MerchantId:   merchantId,
		SessionNo:    strings.TrimSpace(sessionNo),
		EventType:    eventType,
		EventMessage: strings.TrimSpace(eventMessage),
		UserId:       userId,
		AgentId:      agentId,
	})
	if err != nil {
		return err
	}
	if resp.GetBase().GetCode() != 200 {
		return errors.New(resp.GetBase().GetMsg())
	}
	return nil
}

type sendAgentMessagePayload struct {
	MerchantId      int64  `json:"merchantId"`
	AgentId         int64  `json:"agentId"`
	UserId          int64  `json:"userId"`
	SessionNo       string `json:"sessionNo"`
	MessageType     int64  `json:"messageType"`
	Content         string `json:"content"`
	Url             string `json:"url"`
	FileName        string `json:"fileName"`
	MimeType        string `json:"mimeType"`
	FileSize        int64  `json:"fileSize"`
	Width           int32  `json:"width"`
	Height          int32  `json:"height"`
	Duration        int64  `json:"duration"`
	SenderNickname  string `json:"senderNickname"`
	SenderAvatarUrl string `json:"senderAvatarUrl"`
	Extra           string `json:"extra"`
}

type acceptChatSessionPayload struct {
	MerchantId int64  `json:"merchantId"`
	AgentId    int64  `json:"agentId"`
	UserId     int64  `json:"userId"`
	SessionNo  string `json:"sessionNo"`
}

type closeChatSessionPayload struct {
	MerchantId  int64  `json:"merchantId"`
	UserId      int64  `json:"userId"`
	SessionNo   string `json:"sessionNo"`
	CloseReason string `json:"closeReason"`
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}
