package chat_ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"
	"chat-admin-api/internal/ws"
	"wklive/common/helper"
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

	client := ws.NewConnection(l.svcCtx.ChatMessageHub, conn, req.UserId, req.Username, req.MerchantId, req.AgentId, req.SessionNo, l.handleInbound())
	l.svcCtx.ChatMessageHub.Register(client)
	client.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM, map[string]interface{}{
		"message":    "chat admin websocket connected",
		"merchantId": req.MerchantId,
		"agentId":    req.AgentId,
		"sessionNo":  req.SessionNo,
	})

	go client.WritePump()
	client.ReadPump()
}

func (l *MessagesLogic) handleInbound() func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE:
			l.handleSendAgentMessage(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED:
			l.handleAcceptChatSession(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE:
			l.handleCloseChatSession(context.Background(), conn, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_TYPING, chat.ChatEventType_CHAT_EVENT_TYPE_STOP_TYPING:
			l.handleAgentTyping(context.Background(), conn, event.Type, event.Data)
		case chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE:
			l.handleEvaluationInvite(context.Background(), conn, event.Data)
		default:
			sendWSError(conn, "unsupported event type")
		}
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
		Url:         data.MediaUrl,
		FileName:    data.MediaName,
		MimeType:    data.MediaMime,
		FileSize:    data.MediaSize,
	}
	if l.isTransientSession(req.SessionNo) {
		l.fillTransientUserId(conn.MerchantId, req.SessionNo, &data)
		l.fillAgentSenderSnapshot(ctx, conn, &data)
		msg := newTransientAgentMessage(conn.MerchantId, req.SessionNo, data.UserId, conn.AgentId, conn.Username, data)
		if err := publishTransientMessage(ctx, l.svcCtx, msg); err != nil {
			sendWSError(conn, err.Error())
			return
		}
		conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	}

	resp, err := l.svcCtx.ChatAdminCli.SendAgentMessage(ctx, &req)
	if err != nil {
		sendWSError(conn, err.Error())
		return
	}
	conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE, resp)
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
	if l.isTransientSession(sessionNo) {
		data.UserId = l.transientUserId(conn.MerchantId, sessionNo, data.UserId)
		msg := newTransientSystemMessage(conn.MerchantId, sessionNo, data.UserId, agentId, l.agentServiceMessage(ctx, conn))
		if err := publishTransientEvent(ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED, msg); err != nil {
			sendWSError(conn, err.Error())
			return
		}
		conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_ASSIGNED, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	}

	resp, err := l.svcCtx.ChatAdminCli.AcceptChatSession(contextWithAdminIdentity(ctx, conn), &chat.AcceptChatSessionReq{
		SessionNo: sessionNo,
		Reason:    "accept",
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
	if l.isTransientSession(sessionNo) {
		data.UserId = l.transientUserId(conn.MerchantId, sessionNo, data.UserId)
		msg := newTransientSystemMessage(conn.MerchantId, sessionNo, data.UserId, conn.AgentId, "本次会话已结束")
		if err := publishTransientEvent(ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE, msg); err != nil {
			sendWSError(conn, err.Error())
			return
		}
		conn.SendJSON(chat.ChatEventType_CHAT_EVENT_TYPE_SESSION_CLOSE, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	}

	resp, err := l.svcCtx.ChatAdminCli.CloseChatSession(contextWithAdminIdentity(ctx, conn), &chat.CloseChatSessionReq{
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
		userId = l.transientUserId(conn.MerchantId, sessionNo, 0)
	}
	msg := newTransientSystemMessageWithType(conn.MerchantId, sessionNo, userId, conn.AgentId, eventType, chat.ChatMessageType_CHAT_MESSAGE_TYPE_TEXT, message, "")
	if err := publishTransientEvent(ctx, l.svcCtx, eventType, msg); err != nil {
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
		userId = l.transientUserId(conn.MerchantId, sessionNo, 0)
	}
	msg := newTransientSystemMessageWithType(conn.MerchantId, sessionNo, userId, conn.AgentId, chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE, chat.ChatMessageType_CHAT_MESSAGE_TYPE_EVALUATION, content, transientExtra(map[string]interface{}{"action": "invite"}))
	if err := publishTransientEvent(ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_INVITE, msg); err != nil {
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

func (l *MessagesLogic) isTransientSession(sessionNo string) bool {
	return l.svcCtx != nil && l.svcCtx.ChatMessageHub != nil && l.svcCtx.ChatMessageHub.IsTransientSession(strings.TrimSpace(sessionNo))
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

func contextWithAdminIdentity(ctx context.Context, conn *ws.Connection) context.Context {
	ctx = context.WithValue(ctx, utils.CtxKeyUid, conn.UserId)
	ctx = context.WithValue(ctx, utils.CtxKeyUsername, conn.Username)
	ctx = context.WithValue(ctx, utils.CtxKeyMerchantId, conn.MerchantId)
	return ctx
}

func (l *MessagesLogic) agentServiceMessage(ctx context.Context, conn *ws.Connection) string {
	name := ""
	if l.svcCtx != nil && conn != nil && conn.UserId > 0 {
		profileCtx := context.WithValue(ctx, utils.CtxKeyUid, conn.UserId)
		profileCtx = context.WithValue(profileCtx, utils.CtxKeyUsername, conn.Username)
		resp, err := l.svcCtx.ChatAdminCli.Profile(profileCtx, &chat.ChatAdminProfileReq{})
		if err == nil && resp != nil && resp.User != nil {
			name = strings.TrimSpace(resp.User.Nickname)
		}
	}
	if name == "" && conn != nil {
		name = strings.TrimSpace(conn.Username)
	}
	if name == "" {
		return "客服正在为你服务"
	}
	return name + " 客服正在为你服务"
}

func (l *MessagesLogic) fillTransientUserId(merchantId int64, sessionNo string, data *sendAgentMessagePayload) {
	if l.svcCtx == nil || l.svcCtx.ChatMessageHub == nil || data == nil || data.UserId != 0 || strings.TrimSpace(sessionNo) == "" {
		return
	}
	sessions := l.svcCtx.ChatMessageHub.ListTransientSessions(ws.TransientSessionFilter{
		MerchantId: merchantId,
	})
	for _, session := range sessions {
		if session == nil || session.GetSessionNo() != sessionNo {
			continue
		}
		data.UserId = session.GetUserId()
		return
	}
}

func (l *MessagesLogic) transientUserId(merchantId int64, sessionNo string, userId int64) int64 {
	if userId != 0 {
		return userId
	}
	payload := sendAgentMessagePayload{}
	l.fillTransientUserId(merchantId, sessionNo, &payload)
	return payload.UserId
}

func (l *MessagesLogic) fillAgentSenderSnapshot(ctx context.Context, conn *ws.Connection, data *sendAgentMessagePayload) {
	if data == nil || l.svcCtx == nil || conn == nil || conn.UserId <= 0 {
		return
	}
	profileCtx := context.WithValue(ctx, utils.CtxKeyUid, conn.UserId)
	profileCtx = context.WithValue(profileCtx, utils.CtxKeyUsername, conn.Username)
	resp, err := l.svcCtx.ChatAdminCli.Profile(profileCtx, &chat.ChatAdminProfileReq{})
	if err != nil || resp == nil || resp.User == nil {
		return
	}
	if strings.TrimSpace(data.SenderNickname) == "" {
		data.SenderNickname = resp.User.Nickname
	}
	if strings.TrimSpace(data.SenderAvatarUrl) == "" {
		data.SenderAvatarUrl = resp.User.AvatarUrl
	}
}

type sendAgentMessagePayload struct {
	MerchantId      int64  `json:"merchantId"`
	AgentId         int64  `json:"agentId"`
	UserId          int64  `json:"userId"`
	SessionNo       string `json:"sessionNo"`
	MessageType     int64  `json:"messageType"`
	Content         string `json:"content"`
	MediaUrl        string `json:"mediaUrl"`
	MediaName       string `json:"mediaName"`
	MediaMime       string `json:"mediaMime"`
	MediaSize       int64  `json:"mediaSize"`
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
