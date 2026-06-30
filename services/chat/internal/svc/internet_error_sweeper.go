package svc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	internetErrorCloseDelay      = 3 * time.Minute
	internetErrorSweepInterval   = 10 * time.Second
	internetErrorSweepBatchLimit = 100
	transientInternetErrorKey    = "chat:session:internet_error_timeout"
	transientSessionTTLSeconds   = 24 * 60 * 60
)

func (s *ServiceContext) startInternetErrorSessionSweeper() {
	ctx, cancel := context.WithCancel(context.Background())
	s.sweepCancel = cancel
	go s.runInternetErrorSessionSweeper(ctx)
}

func (s *ServiceContext) runInternetErrorSessionSweeper(ctx context.Context) {
	ticker := time.NewTicker(internetErrorSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.sweepExpiredInternetErrorSessions(ctx); err != nil {
				logx.Errorf("sweep expired internet error chat sessions failed: %v", err)
			}
		}
	}
}

func (s *ServiceContext) sweepExpiredInternetErrorSessions(ctx context.Context) error {
	now := utils.NowMillis()
	beforeMillis := now - int64(internetErrorCloseDelay/time.Millisecond)
	list, err := s.ChatSessionModel.FindExpiredInternetError(ctx, beforeMillis, internetErrorSweepBatchLimit)
	if err != nil {
		return err
	}
	for _, session := range list {
		current, err := s.ChatSessionModel.FindOneBySessionNo(ctx, session.SessionNo)
		if err != nil {
			if err == models.ErrNotFound {
				continue
			}
			return err
		}
		if !isExpiredInternetErrorSession(current, now) {
			continue
		}
		if err := s.closeExpiredInternetErrorSession(ctx, current, now); err != nil {
			return err
		}
	}
	return s.sweepExpiredTransientInternetErrorSessions(ctx, now)
}

func (s *ServiceContext) closeExpiredInternetErrorSession(ctx context.Context, session *models.TChatSession, now int64) error {
	oldAgentID := session.AgentId
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED)
	session.CloseTime = now
	session.CloseReason = "用户网络异常超时关闭"
	session.DisconnectTime = 0
	session.BeforeDisconnectStatus = 0
	session.UpdateTimes = now
	if err := s.ChatSessionModel.Update(ctx, session); err != nil {
		return err
	}
	return s.changeAgentSessionCount(ctx, oldAgentID, -1)
}

func (s *ServiceContext) changeAgentSessionCount(ctx context.Context, agentID int64, delta int64) error {
	if agentID <= 0 || delta == 0 {
		return nil
	}
	agent, err := s.ChatAgentModel.FindOne(ctx, agentID)
	if err == models.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	agent.CurrentSessionCount += delta
	if agent.CurrentSessionCount < 0 {
		agent.CurrentSessionCount = 0
	}
	agent.UpdateTimes = utils.NowMillis()
	return s.ChatAgentModel.Update(ctx, agent)
}

func isExpiredInternetErrorSession(session *models.TChatSession, now int64) bool {
	if session == nil || session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
		return false
	}
	return session.DisconnectTime > 0 && now-session.DisconnectTime >= int64(internetErrorCloseDelay/time.Millisecond)
}

func (s *ServiceContext) sweepExpiredTransientInternetErrorSessions(ctx context.Context, now int64) error {
	if s.BusRedis == nil {
		return nil
	}
	pairs, err := s.BusRedis.ZrangebyscoreWithScoresAndLimitCtx(ctx, transientInternetErrorKey, 0, now, 0, internetErrorSweepBatchLimit)
	if err != nil {
		return err
	}
	for _, pair := range pairs {
		merchantID, sessionNo, ok := parseTransientInternetErrorMember(pair.Key)
		if !ok {
			_, _ = s.BusRedis.ZremCtx(ctx, transientInternetErrorKey, pair.Key)
			continue
		}
		session, err := s.getTransientSession(ctx, merchantID, sessionNo)
		if err != nil || session == nil {
			_, _ = s.BusRedis.ZremCtx(ctx, transientInternetErrorKey, pair.Key)
			continue
		}
		if session.GetStatus() != chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR {
			_, _ = s.BusRedis.ZremCtx(ctx, transientInternetErrorKey, pair.Key)
			continue
		}
		if !isExpiredTransientInternetErrorSession(session, now) {
			continue
		}
		session.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED
		session.CloseTime = now
		session.CloseReason = "用户网络异常超时关闭"
		session.DisconnectTime = 0
		session.BeforeDisconnectStatus = chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN
		session.UpdateTimes = now
		if err := s.upsertTransientSession(ctx, session); err != nil {
			return err
		}
		_, _ = s.BusRedis.ZremCtx(ctx, transientInternetErrorKey, pair.Key)
	}
	return nil
}

func (s *ServiceContext) getTransientSession(ctx context.Context, merchantID int64, sessionNo string) (*chat.ChatSession, error) {
	if s.BusRedis == nil {
		return nil, nil
	}
	data, err := s.BusRedis.GetCtx(ctx, transientSessionKey(merchantID, sessionNo))
	if err != nil || strings.TrimSpace(data) == "" {
		return nil, err
	}
	var session chat.ChatSession
	if err := protojson.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *ServiceContext) upsertTransientSession(ctx context.Context, session *chat.ChatSession) error {
	if s.BusRedis == nil || session == nil {
		return nil
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(session)
	if err != nil {
		return err
	}
	sessionKey := transientSessionKey(session.GetMerchantId(), session.GetSessionNo())
	indexKey := transientSessionIndexKey(session.GetMerchantId())
	if err := s.BusRedis.SetexCtx(ctx, sessionKey, string(payload), transientSessionTTLSeconds); err != nil {
		return err
	}
	if _, err := s.BusRedis.ZaddCtx(ctx, indexKey, transientSessionSortScore(session), session.GetSessionNo()); err != nil {
		return err
	}
	_ = s.BusRedis.ExpireCtx(ctx, indexKey, transientSessionTTLSeconds)
	return nil
}

func isExpiredTransientInternetErrorSession(session *chat.ChatSession, now int64) bool {
	if session == nil || session.GetStatus() != chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR {
		return false
	}
	return session.GetDisconnectTime() > 0 && now-session.GetDisconnectTime() >= int64(internetErrorCloseDelay/time.Millisecond)
}

func transientSessionSortScore(session *chat.ChatSession) int64 {
	return maxInt64(session.GetLastMessageTime(), session.GetUpdateTimes(), session.GetCreateTimes())
}

func transientSessionIndexKey(merchantID int64) string {
	return fmt.Sprintf("chat:transient:sessions:%d", merchantID)
}

func transientSessionKey(merchantID int64, sessionNo string) string {
	return fmt.Sprintf("chat:transient:session:%d:%s", merchantID, strings.TrimSpace(sessionNo))
}

func parseTransientInternetErrorMember(member string) (int64, string, bool) {
	merchant, sessionNo, ok := strings.Cut(member, ":")
	if !ok || strings.TrimSpace(sessionNo) == "" {
		return 0, "", false
	}
	merchantID, err := strconv.ParseInt(merchant, 10, 64)
	if err != nil || merchantID <= 0 {
		return 0, "", false
	}
	return merchantID, sessionNo, true
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
