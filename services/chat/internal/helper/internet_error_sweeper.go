package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	internetErrorSweepInterval   = 10 * time.Second
	internetErrorSweepBatchLimit = 100
)

func StartInternetErrorSessionSweeper(svcCtx *svc.ServiceContext) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go runInternetErrorSessionSweeper(ctx, svcCtx)
	return cancel
}

func runInternetErrorSessionSweeper(ctx context.Context, svcCtx *svc.ServiceContext) {
	ticker := time.NewTicker(internetErrorSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sweepExpiredInternetErrorSessions(ctx, svcCtx); err != nil {
				logx.Errorf("sweep expired internet error chat sessions failed: %v", err)
			}
		}
	}
}

func sweepExpiredInternetErrorSessions(ctx context.Context, svcCtx *svc.ServiceContext) error {
	now := utils.NowMillis()
	beforeMillis := now - int64(InternetErrorCloseDelay/time.Millisecond)
	list, err := svcCtx.ChatSessionModel.FindExpiredInternetError(ctx, beforeMillis, internetErrorSweepBatchLimit)
	if err != nil {
		return err
	}
	for _, session := range list {
		current, err := svcCtx.ChatSessionModel.FindOneBySessionNo(ctx, session.SessionNo)
		if err != nil {
			if err == models.ErrNotFound {
				continue
			}
			return err
		}
		if !IsInternetErrorExpired(current, now) {
			continue
		}
		if err := CloseSession(ctx, svcCtx, current, "用户网络异常超时关闭"); err != nil {
			return err
		}
	}
	return sweepExpiredTransientInternetErrorSessions(ctx, svcCtx, now)
}

func sweepExpiredTransientInternetErrorSessions(ctx context.Context, svcCtx *svc.ServiceContext, now int64) error {
	if svcCtx.BusRedis == nil {
		return nil
	}
	pairs, err := svcCtx.BusRedis.ZrangebyscoreWithScoresAndLimitCtx(ctx, transientInternetErrorKey, 0, now, 0, internetErrorSweepBatchLimit)
	if err != nil {
		return err
	}
	for _, pair := range pairs {
		merchantID, sessionNo, ok := parseTransientInternetErrorMember(pair.Key)
		if !ok {
			_, _ = svcCtx.BusRedis.ZremCtx(ctx, transientInternetErrorKey, pair.Key)
			continue
		}
		session, err := GetTransientSession(ctx, svcCtx.BusRedis, merchantID, sessionNo)
		if err != nil || session == nil {
			_, _ = svcCtx.BusRedis.ZremCtx(ctx, transientInternetErrorKey, pair.Key)
			continue
		}
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
			_, _ = svcCtx.BusRedis.ZremCtx(ctx, transientInternetErrorKey, pair.Key)
			continue
		}
		if !IsInternetErrorExpired(session, now) {
			continue
		}
		session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED)
		session.CloseTime = now
		session.CloseReason = "用户网络异常超时关闭"
		session.DisconnectTime = 0
		session.BeforeDisconnectStatus = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN)
		session.UpdateTimes = now
		if _, err := UpsertTransientSession(ctx, svcCtx.BusRedis, session); err != nil {
			return err
		}
		_, _ = svcCtx.BusRedis.ZremCtx(ctx, transientInternetErrorKey, pair.Key)
	}
	return nil
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
