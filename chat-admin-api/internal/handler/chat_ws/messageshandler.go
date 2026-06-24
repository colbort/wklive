package chat_ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/ws"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	eventConnected               = chat.ChatWsEventConnected
	eventError                   = chat.ChatWsEventError
	eventSendAgentMessage        = chat.ChatWsEventSendAgentMessage
	eventSendAgentMessageResult  = chat.ChatWsEventSendAgentMessageResult
	eventAcceptChatSession       = chat.ChatWsEventAcceptChatSession
	eventAcceptChatSessionResult = chat.ChatWsEventAcceptChatSessionResult
	eventCloseChatSession        = chat.ChatWsEventCloseChatSession
	eventCloseChatSessionResult  = chat.ChatWsEventCloseChatSessionResult
	eventChatSessionAccepted     = chat.ChatMessageEventTypeSessionAccepted
	eventChatSessionClosed       = chat.ChatMessageEventTypeSessionClosed
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := parseToken(r, svcCtx.Config.Jwt.AccessSecret)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		merchantId := parseInt64(r.URL.Query().Get("merchantId"))
		agentId := parseInt64(r.URL.Query().Get("agentId"))
		sessionNo := strings.TrimSpace(r.URL.Query().Get("sessionNo"))

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("upgrade chat admin ws failed, userId=%d err=%v", claims.UserId, err)
			return
		}

		client := ws.NewConnection(svcCtx.ChatMessageHub, conn, claims.UserId, claims.Username, merchantId, agentId, sessionNo, handleInbound(svcCtx))
		svcCtx.ChatMessageHub.Register(client)
		client.SendJSON(eventConnected, map[string]interface{}{
			"message":    "chat admin websocket connected",
			"merchantId": merchantId,
			"agentId":    agentId,
			"sessionNo":  sessionNo,
		})

		go client.WritePump()
		client.ReadPump()
	}
}

func handleInbound(svcCtx *svc.ServiceContext) func(*ws.Connection, ws.InboundEvent) {
	return func(conn *ws.Connection, event ws.InboundEvent) {
		switch event.Type {
		case eventSendAgentMessage:
			handleSendAgentMessage(context.Background(), svcCtx, conn, event.Data)
		case eventAcceptChatSession:
			handleAcceptChatSession(context.Background(), svcCtx, conn, event.Data)
		case eventCloseChatSession:
			handleCloseChatSession(context.Background(), svcCtx, conn, event.Data)
		default:
			conn.SendJSON(eventError, map[string]string{"message": "unsupported event type"})
		}
	}
}

func handleSendAgentMessage(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, payload json.RawMessage) {
	var data sendAgentMessagePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		conn.SendJSON(eventError, map[string]string{"message": "invalid send_agent_message payload"})
		return
	}
	req := chat.SendAgentMessageReq{
		AgentId:     data.AgentId,
		SessionNo:   data.SessionNo,
		MessageType: chat.ChatMessageType(data.MessageType),
		Content:     data.Content,
		MediaUrl:    data.MediaUrl,
		MediaName:   data.MediaName,
		MediaMime:   data.MediaMime,
		MediaSize:   data.MediaSize,
	}
	if req.AgentId == 0 {
		req.AgentId = conn.AgentId
	}
	if req.SessionNo == "" {
		req.SessionNo = conn.SessionNo
	}
	if isGuestSession(req.SessionNo) {
		fillTransientUserId(svcCtx, conn.MerchantId, req.SessionNo, &data)
		fillAgentSenderSnapshot(ctx, svcCtx, conn, &data)
		msg := newTransientAgentMessage(conn.MerchantId, req.SessionNo, data.UserId, req.AgentId, conn.Username, data)
		if err := publishTransientMessage(ctx, svcCtx, msg); err != nil {
			conn.SendJSON(eventError, map[string]string{"message": err.Error()})
			return
		}
		conn.SendJSON(eventSendAgentMessageResult, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	}

	resp, err := svcCtx.ChatAdminCli.SendAgentMessage(ctx, &req)
	if err != nil {
		conn.SendJSON(eventError, map[string]string{"message": err.Error()})
		return
	}
	conn.SendJSON(eventSendAgentMessageResult, resp)
}

func handleAcceptChatSession(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, payload json.RawMessage) {
	var data acceptChatSessionPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		conn.SendJSON(eventError, map[string]string{"message": "invalid accept_chat_session payload"})
		return
	}
	sessionNo := strings.TrimSpace(data.SessionNo)
	if sessionNo == "" {
		sessionNo = conn.SessionNo
	}
	agentId := data.AgentId
	if agentId == 0 {
		agentId = conn.AgentId
	}
	if sessionNo == "" || agentId == 0 {
		conn.SendJSON(eventError, map[string]string{"message": "sessionNo and agentId are required"})
		return
	}
	if isGuestSession(sessionNo) {
		msg := newTransientSystemMessage(conn.MerchantId, sessionNo, data.UserId, agentId, "客服已接入")
		if err := publishTransientEvent(ctx, svcCtx, eventChatSessionAccepted, msg); err != nil {
			conn.SendJSON(eventError, map[string]string{"message": err.Error()})
			return
		}
		conn.SendJSON(eventAcceptChatSessionResult, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	}

	resp, err := svcCtx.ChatAdminCli.AcceptChatSession(contextWithAdminIdentity(ctx, conn), &chat.AcceptChatSessionReq{
		SessionNo:  sessionNo,
		AgentId:    agentId,
		OperatorId: conn.UserId,
		Reason:     "accept",
	})
	if err != nil {
		conn.SendJSON(eventError, map[string]string{"message": err.Error()})
		return
	}
	conn.SendJSON(eventAcceptChatSessionResult, resp)
}

func handleCloseChatSession(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, payload json.RawMessage) {
	var data closeChatSessionPayload
	if err := json.Unmarshal(payload, &data); err != nil {
		conn.SendJSON(eventError, map[string]string{"message": "invalid close_chat_session payload"})
		return
	}
	sessionNo := strings.TrimSpace(data.SessionNo)
	if sessionNo == "" {
		sessionNo = conn.SessionNo
	}
	if sessionNo == "" {
		conn.SendJSON(eventError, map[string]string{"message": "sessionNo is required"})
		return
	}
	reason := firstNonEmpty(data.CloseReason, "closed by agent")
	if isGuestSession(sessionNo) {
		msg := newTransientSystemMessage(conn.MerchantId, sessionNo, data.UserId, conn.AgentId, "本次会话已结束")
		if err := publishTransientEvent(ctx, svcCtx, eventChatSessionClosed, msg); err != nil {
			conn.SendJSON(eventError, map[string]string{"message": err.Error()})
			return
		}
		conn.SendJSON(eventCloseChatSessionResult, &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg})
		return
	}

	resp, err := svcCtx.ChatAdminCli.CloseChatSession(contextWithAdminIdentity(ctx, conn), &chat.CloseChatSessionReq{
		SessionNo:   sessionNo,
		CloseReason: reason,
	})
	if err != nil {
		conn.SendJSON(eventError, map[string]string{"message": err.Error()})
		return
	}
	conn.SendJSON(eventCloseChatSessionResult, resp)
}

func contextWithAdminIdentity(ctx context.Context, conn *ws.Connection) context.Context {
	ctx = context.WithValue(ctx, utils.CtxKeyUid, conn.UserId)
	ctx = context.WithValue(ctx, utils.CtxKeyUsername, conn.Username)
	ctx = context.WithValue(ctx, utils.CtxKeyMerchantId, conn.MerchantId)
	return ctx
}

func fillTransientUserId(svcCtx *svc.ServiceContext, merchantId int64, sessionNo string, data *sendAgentMessagePayload) {
	if svcCtx == nil || svcCtx.ChatMessageHub == nil || data == nil || data.UserId != 0 || strings.TrimSpace(sessionNo) == "" {
		return
	}
	sessions := svcCtx.ChatMessageHub.ListTransientSessions(ws.TransientSessionFilter{
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

func fillAgentSenderSnapshot(ctx context.Context, svcCtx *svc.ServiceContext, conn *ws.Connection, data *sendAgentMessagePayload) {
	if data == nil || svcCtx == nil || conn == nil || conn.UserId <= 0 {
		return
	}
	profileCtx := context.WithValue(ctx, utils.CtxKeyUid, conn.UserId)
	profileCtx = context.WithValue(profileCtx, utils.CtxKeyUsername, conn.Username)
	resp, err := svcCtx.ChatAdminCli.Profile(profileCtx, &chat.ChatAdminProfileReq{})
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

func parseToken(r *http.Request, secret string) (*utils.Claims, error) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("Sec-WebSocket-Protocol"))
	}
	if token == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return utils.ParseToken(secret, token)
}

func parseInt64(value string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return n
}
