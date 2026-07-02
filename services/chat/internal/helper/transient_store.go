package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"wklive/proto/chat"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	DefaultTransientTTLSeconds = int(24 * 60 * 60)
	transientMessageLimit      = int64(200)
)

func UpsertTransientSession(ctx context.Context, rds *redis.Redis, session *models.TChatSession) (*models.TChatSession, error) {
	if rds == nil {
		return nil, fmt.Errorf("chat redis is not configured")
	}
	if session.MerchantId <= 0 || strings.TrimSpace(session.SessionNo) == "" {
		return nil, fmt.Errorf("merchant_id and session_no are required")
	}
	payload, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	sessionKey := transientSessionKey(session.MerchantId, session.SessionNo)
	indexKey := transientSessionIndexKey(session.MerchantId)
	if err := rds.SetexCtx(ctx, sessionKey, string(payload), DefaultTransientTTLSeconds); err != nil {
		return nil, err
	}
	if _, err := rds.ZaddCtx(ctx, indexKey, sessionSortScore(session), session.SessionNo); err != nil {
		return nil, err
	}
	_ = rds.ExpireCtx(ctx, indexKey, DefaultTransientTTLSeconds)
	return session, nil
}

func GetTransientSession(ctx context.Context, rds *redis.Redis, merchantId int64, sessionNo string) (*models.TChatSession, error) {
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
	var session models.TChatSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}
	return &session, nil
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

func PageTransientSessions(ctx context.Context, rds *redis.Redis, merchantId, userId, agentId, status, cursor, limit int64) ([]*models.TChatSession, bool, int64, error) {
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
	filtered := make([]*models.TChatSession, 0, len(nos))
	for _, sessionNo := range nos {
		session, err := GetTransientSession(ctx, rds, merchantId, sessionNo)
		if err != nil || session == nil {
			_, _ = rds.ZremCtx(ctx, transientSessionIndexKey(merchantId), sessionNo)
			continue
		}
		if userId > 0 && session.UserId != userId {
			continue
		}
		if agentId > 0 && session.AgentId != agentId {
			continue
		}
		if status > 0 && !transientSessionStatusMatches(status, session.Status) {
			continue
		}
		filtered = append(filtered, session)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return sessionSortScore(filtered[i]) > sessionSortScore(filtered[j])
	})
	if cursor >= int64(len(filtered)) {
		return []*models.TChatSession{}, false, cursor, nil
	}
	end := cursor + limit
	if end > int64(len(filtered)) {
		end = int64(len(filtered))
	}
	return filtered[cursor:end], end < int64(len(filtered)), end, nil
}

func AppendTransientMessage(ctx context.Context, rds *redis.Redis, merchantId int64, msg *chat.ChatMessage, session *models.TChatSession) (*chat.ChatMessage, error) {
	if rds == nil {
		return nil, fmt.Errorf("chat redis is not configured")
	}
	if msg == nil || strings.TrimSpace(msg.GetSessionNo()) == "" {
		return nil, fmt.Errorf("message.session_no is required")
	}
	if merchantId <= 0 && session != nil {
		merchantId = session.MerchantId
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
	msgKey := transientMessagesKey(merchantId, msg.GetSessionNo())
	if _, err := rds.LpushCtx(ctx, msgKey, string(payload)); err != nil {
		return nil, err
	}
	_ = rds.LtrimCtx(ctx, msgKey, 0, transientMessageLimit-1)
	_ = rds.ExpireCtx(ctx, msgKey, DefaultTransientTTLSeconds)
	if _, err := UpsertTransientSession(ctx, rds, session); err != nil {
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

func UpdateTransientMessageStatus(ctx context.Context, rds *redis.Redis, merchantId int64, sessionNo, messageNo string, status chat.ChatMessageStatus, updateTimes int64) error {
	if rds == nil {
		return fmt.Errorf("chat redis is not configured")
	}
	sessionNo = strings.TrimSpace(sessionNo)
	messageNo = strings.TrimSpace(messageNo)
	if merchantId <= 0 || sessionNo == "" || messageNo == "" {
		return fmt.Errorf("merchant_id, session_no and message_no are required")
	}
	key := transientMessagesKey(merchantId, sessionNo)
	raw, err := rds.LrangeCtx(ctx, key, 0, int(transientMessageLimit-1))
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return fmt.Errorf("chat message not found")
	}
	updated := false
	next := make([]string, 0, len(raw))
	for _, item := range raw {
		var msg chat.ChatMessage
		if err := protojson.Unmarshal([]byte(item), &msg); err != nil {
			next = append(next, item)
			continue
		}
		if msg.GetMessageNo() == messageNo {
			msg.Status = status
			msg.UpdateTimes = updateTimes
			payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(&msg)
			if err != nil {
				return err
			}
			next = append(next, string(payload))
			updated = true
			continue
		}
		next = append(next, item)
	}
	if !updated {
		return fmt.Errorf("chat message not found")
	}
	if _, err := rds.DelCtx(ctx, key); err != nil {
		return err
	}
	for i := len(next) - 1; i >= 0; i-- {
		if _, err := rds.LpushCtx(ctx, key, next[i]); err != nil {
			return err
		}
	}
	return rds.ExpireCtx(ctx, key, DefaultTransientTTLSeconds)
}

func transientSessionStatusMatches(filter int64, status int64) bool {
	if filter == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) {
		return status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) ||
			status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) ||
			status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT)
	}
	return status == filter
}

func sessionSortScore(session *models.TChatSession) int64 {
	return maxInt64(session.LastMessageTime, session.UpdateTimes, session.CreateTimes)
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
