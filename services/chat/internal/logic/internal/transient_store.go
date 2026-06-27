package internal

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	DefaultTransientTTLSeconds = int64(24 * 60 * 60)
	transientMessageLimit      = int64(200)
)

func UpsertTransientSession(ctx context.Context, rds *redis.Redis, session *chat.ChatSession, ttlSeconds int64) (*chat.ChatSession, error) {
	if rds == nil {
		return nil, fmt.Errorf("chat redis is not configured")
	}
	session = normalizeTransientSession(session)
	if session.GetMerchantId() <= 0 || strings.TrimSpace(session.GetSessionNo()) == "" {
		return nil, fmt.Errorf("merchant_id and session_no are required")
	}
	ttl := normalizeTransientTTL(ttlSeconds)
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(session)
	if err != nil {
		return nil, err
	}
	sessionKey := transientSessionKey(session.GetMerchantId(), session.GetSessionNo())
	indexKey := transientSessionIndexKey(session.GetMerchantId())
	if err := rds.SetexCtx(ctx, sessionKey, string(payload), int(ttl)); err != nil {
		return nil, err
	}
	if _, err := rds.ZaddCtx(ctx, indexKey, sessionSortScore(session), session.GetSessionNo()); err != nil {
		return nil, err
	}
	_ = rds.ExpireCtx(ctx, indexKey, int(ttl))
	return session, nil
}

func GetTransientSession(ctx context.Context, rds *redis.Redis, merchantId int64, sessionNo string) (*chat.ChatSession, error) {
	if rds == nil {
		return nil, fmt.Errorf("chat redis is not configured")
	}
	sessionNo = strings.TrimSpace(sessionNo)
	if merchantId <= 0 || sessionNo == "" {
		return nil, fmt.Errorf("merchant_id and session_no are required")
	}
	data, err := rds.GetCtx(ctx, transientSessionKey(merchantId, sessionNo))
	if err != nil || strings.TrimSpace(data) == "" {
		return nil, err
	}
	var session chat.ChatSession
	if err := protojson.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}
	return normalizeTransientSession(&session), nil
}

func DeleteTransientSession(ctx context.Context, rds *redis.Redis, merchantId int64, sessionNo string) error {
	if rds == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	sessionNo = strings.TrimSpace(sessionNo)
	if merchantId <= 0 || sessionNo == "" {
		return fmt.Errorf("merchant_id and session_no are required")
	}
	_, err := rds.DelCtx(ctx, transientSessionKey(merchantId, sessionNo), transientMessagesKey(merchantId, sessionNo))
	if err != nil {
		return err
	}
	_, err = rds.ZremCtx(ctx, transientSessionIndexKey(merchantId), sessionNo)
	return err
}

func PageTransientSessions(ctx context.Context, rds *redis.Redis, merchantId, userId, agentId, status, cursor, limit int64) ([]*chat.ChatSession, bool, int64, error) {
	if rds == nil {
		return nil, false, 0, fmt.Errorf("chat redis is not configured")
	}
	if merchantId <= 0 {
		return nil, false, 0, fmt.Errorf("merchant_id is required")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if cursor < 0 {
		cursor = 0
	}
	nos, err := rds.ZrevrangeCtx(ctx, transientSessionIndexKey(merchantId), 0, -1)
	if err != nil {
		return nil, false, 0, err
	}
	filtered := make([]*chat.ChatSession, 0, len(nos))
	for _, sessionNo := range nos {
		session, err := GetTransientSession(ctx, rds, merchantId, sessionNo)
		if err != nil || session == nil {
			_, _ = rds.ZremCtx(ctx, transientSessionIndexKey(merchantId), sessionNo)
			continue
		}
		if userId > 0 && session.GetUserId() != userId {
			continue
		}
		if agentId > 0 && session.GetAgentId() != agentId {
			continue
		}
		if status > 0 && !transientSessionStatusMatches(status, session.GetStatus()) {
			continue
		}
		filtered = append(filtered, session)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return sessionSortScore(filtered[i]) > sessionSortScore(filtered[j])
	})
	if cursor >= int64(len(filtered)) {
		return []*chat.ChatSession{}, false, cursor, nil
	}
	end := cursor + limit
	if end > int64(len(filtered)) {
		end = int64(len(filtered))
	}
	return filtered[cursor:end], end < int64(len(filtered)), end, nil
}

func AppendTransientMessage(ctx context.Context, rds *redis.Redis, merchantId int64, msg *chat.ChatMessage, session *chat.ChatSession, ttlSeconds int64) (*chat.ChatMessage, error) {
	if rds == nil {
		return nil, fmt.Errorf("chat redis is not configured")
	}
	if msg == nil || strings.TrimSpace(msg.GetSessionNo()) == "" {
		return nil, fmt.Errorf("message.session_no is required")
	}
	if merchantId <= 0 {
		merchantId = session.GetMerchantId()
	}
	if merchantId <= 0 {
		return nil, fmt.Errorf("merchant_id is required")
	}
	now := time.Now().UnixMilli()
	if msg.GetCreateTimes() == 0 {
		msg.CreateTimes = now
	}
	if msg.GetUpdateTimes() == 0 {
		msg.UpdateTimes = msg.GetCreateTimes()
	}
	if msg.GetStatus() == chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_UNKNOWN {
		msg.Status = chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_SENT
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(msg)
	if err != nil {
		return nil, err
	}
	ttl := normalizeTransientTTL(ttlSeconds)
	msgKey := transientMessagesKey(merchantId, msg.GetSessionNo())
	if _, err := rds.LpushCtx(ctx, msgKey, string(payload)); err != nil {
		return nil, err
	}
	_ = rds.LtrimCtx(ctx, msgKey, 0, transientMessageLimit-1)
	_ = rds.ExpireCtx(ctx, msgKey, int(ttl))
	nextSession := normalizeTransientSession(session)
	if nextSession.GetSessionNo() == "" {
		nextSession, _ = GetTransientSession(ctx, rds, merchantId, msg.GetSessionNo())
	}
	if nextSession == nil {
		nextSession = &chat.ChatSession{MerchantId: merchantId, SessionNo: msg.GetSessionNo(), IsGuest: true}
	}
	applyTransientMessageToSession(nextSession, msg)
	if _, err := UpsertTransientSession(ctx, rds, nextSession, ttl); err != nil {
		return nil, err
	}
	return msg, nil
}

func ListTransientMessages(ctx context.Context, rds *redis.Redis, merchantId int64, sessionNo string, senderType, cursor, limit int64) ([]*chat.ChatMessage, bool, int64, error) {
	if rds == nil {
		return nil, false, 0, fmt.Errorf("chat redis is not configured")
	}
	if merchantId <= 0 || strings.TrimSpace(sessionNo) == "" {
		return nil, false, 0, fmt.Errorf("merchant_id and session_no are required")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if cursor < 0 {
		cursor = 0
	}
	raw, err := rds.LrangeCtx(ctx, transientMessagesKey(merchantId, sessionNo), 0, int(transientMessageLimit-1))
	if err != nil {
		return nil, false, 0, err
	}
	list := make([]*chat.ChatMessage, 0, len(raw))
	for i := len(raw) - 1; i >= 0; i-- {
		var msg chat.ChatMessage
		if err := protojson.Unmarshal([]byte(raw[i]), &msg); err != nil {
			continue
		}
		if senderType > 0 && int64(msg.GetSender().GetType()) != senderType {
			continue
		}
		list = append(list, &msg)
	}
	if cursor >= int64(len(list)) {
		return []*chat.ChatMessage{}, false, cursor, nil
	}
	end := cursor + limit
	if end > int64(len(list)) {
		end = int64(len(list))
	}
	return list[cursor:end], end < int64(len(list)), end, nil
}

func normalizeTransientSession(session *chat.ChatSession) *chat.ChatSession {
	if session == nil {
		return nil
	}
	now := time.Now().UnixMilli()
	if session.GetCreateTimes() == 0 {
		session.CreateTimes = now
	}
	if session.GetUpdateTimes() == 0 {
		session.UpdateTimes = now
	}
	if session.GetSource() == chat.ChatSessionSource_CHAT_SESSION_SOURCE_UNKNOWN {
		session.Source = chat.ChatSessionSource_CHAT_SESSION_SOURCE_WEB
	}
	if session.GetStatus() == chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN {
		session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING
	}
	if session.GetPriority() == chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_UNKNOWN {
		session.Priority = chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL
	}
	session.IsGuest = true
	return session
}

func applyTransientMessageToSession(session *chat.ChatSession, msg *chat.ChatMessage) {
	if session == nil || msg == nil {
		return
	}
	session.LastMessage = msg.GetContent()
	session.LastMessageNo = msg.GetMessageNo()
	session.LastSenderType = msg.GetSender().GetType()
	session.LastMessageTime = msg.GetCreateTimes()
	session.UpdateTimes = msg.GetUpdateTimes()
	if session.GetUserId() == 0 && msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
		session.UserId = msg.GetSender().GetId()
	}
	if session.GetAgentId() == 0 && msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		session.AgentId = msg.GetSender().GetId()
	}
	if msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
		session.AgentUnreadCount++
		if session.GetAgentId() > 0 {
			session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT
		}
	}
	if msg.GetSender().GetType() == chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT {
		session.UserUnreadCount++
		if session.GetStatus() != chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED {
			session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER
		}
	}
}

func normalizeTransientTTL(ttlSeconds int64) int64 {
	if ttlSeconds <= 0 {
		return DefaultTransientTTLSeconds
	}
	return ttlSeconds
}

func transientSessionStatusMatches(filter int64, status chat.ChatSessionStatus) bool {
	if filter == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) {
		return status == chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING ||
			status == chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER ||
			status == chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT
	}
	return int64(status) == filter
}

func sessionSortScore(session *chat.ChatSession) int64 {
	return maxInt64(session.GetLastMessageTime(), session.GetUpdateTimes(), session.GetCreateTimes())
}

func transientSessionIndexKey(merchantId int64) string {
	return fmt.Sprintf("chat:transient:sessions:%d", merchantId)
}

func transientSessionKey(merchantId int64, sessionNo string) string {
	return fmt.Sprintf("chat:transient:session:%d:%s", merchantId, strings.TrimSpace(sessionNo))
}

func transientMessagesKey(merchantId int64, sessionNo string) string {
	return fmt.Sprintf("chat:transient:messages:%d:%s", merchantId, strings.TrimSpace(sessionNo))
}

func maxInt64(values ...int64) int64 {
	var max int64
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}
