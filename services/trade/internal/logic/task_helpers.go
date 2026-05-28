package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	tradeTaskLockTTL           = 30 * time.Second
	tradeTaskLockRenewInterval = 10 * time.Second
)

var errTradeTaskLockNotAcquired = errors.New("trade task lock not acquired")

func okTradeTaskResp() *trade.TradeTaskResp {
	return &trade.TradeTaskResp{Base: helper.OkResp()}
}

func runTradeTaskWithLock(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	taskName string,
	fn func() (*trade.TradeTaskResp, error),
) (*trade.TradeTaskResp, error) {
	lockKey := fmt.Sprintf("trade:task:%s", taskName)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	if err := acquireTradeTaskLock(ctx, svcCtx.Redis, lockKey, lockValue); err != nil {
		if errors.Is(err, errTradeTaskLockNotAcquired) {
			return &trade.TradeTaskResp{
				Base: helper.GetErrResp(1, i18n.Translate(i18n.SyncTaskAlreadyRunning, ctx)),
			}, nil
		}
		logx.Errorf("acquire trade task lock failed, key=%s err=%v", lockKey, err)
		return &trade.TradeTaskResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.DistributedLockAcquireFailed, ctx)),
		}, nil
	}

	renewCtx, renewCancel := context.WithCancel(ctx)
	renewDone := make(chan struct{})
	go func() {
		defer close(renewDone)
		autoRenewTradeTaskLock(renewCtx, svcCtx.Redis, lockKey, lockValue)
	}()
	defer func() {
		renewCancel()
		<-renewDone
		if err := releaseTradeTaskLock(context.Background(), svcCtx.Redis, lockKey, lockValue); err != nil {
			logx.Errorf("release trade task lock failed, key=%s err=%v", lockKey, err)
		}
	}()

	return fn()
}

func acquireTradeTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
	ok, err := rds.SetnxExCtx(ctx, key, value, int(tradeTaskLockTTL.Seconds()))
	if err != nil {
		return err
	}
	if !ok {
		return errTradeTaskLockNotAcquired
	}
	return nil
}

func autoRenewTradeTaskLock(ctx context.Context, rds *redis.Redis, key, value string) {
	ticker := time.NewTicker(tradeTaskLockRenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := refreshTradeTaskLock(context.Background(), rds, key, value); err != nil {
				logx.Errorf("refresh trade task lock failed, key=%s err=%v", key, err)
				return
			}
		}
	}
}

func refreshTradeTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
	return 0
end
`
	ret, err := rds.EvalCtx(ctx, lua, []string{key}, value, tradeTaskLockTTL.Milliseconds())
	if err != nil {
		return err
	}
	if !redisEvalOK(ret) {
		return errTradeTaskLockNotAcquired
	}
	return nil
}

func releaseTradeTaskLock(ctx context.Context, rds *redis.Redis, key, value string) error {
	lua := `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`
	_, err := rds.EvalCtx(ctx, lua, []string{key}, value)
	return err
}

func redisEvalOK(ret any) bool {
	switch v := ret.(type) {
	case int64:
		return v > 0
	case int:
		return v > 0
	case int32:
		return v > 0
	case uint64:
		return v > 0
	default:
		return false
	}
}

func createTradeTaskEvent(ctx context.Context, svcCtx *svc.ServiceContext, tenantID int64, eventType, bizType string, bizID int64, userID, symbolID, marketType int64, payload string) error {
	bizIDText := strconv.FormatInt(bizID, 10)
	exists, _, err := svcCtx.BizTradeEventModel.FindPage(ctx, models.BizTradeEventPageFilter{
		TenantId:    tenantID,
		EventType:   eventType,
		BizType:     bizType,
		BizId:       bizIDText,
		EventStatus: int64(trade.EventStatus_EVENT_STATUS_PENDING),
	}, 0, 1)
	if err != nil {
		return err
	}
	if len(exists) > 0 {
		return nil
	}
	eventNo, err := svcCtx.GenerateBizNo(ctx, "TRE")
	if err != nil {
		return err
	}
	now := utils.NowMillis()
	_, err = svcCtx.BizTradeEventModel.Insert(ctx, &models.TBizTradeEvent{
		TenantId:      tenantID,
		EventNo:       eventNo,
		EventType:     eventType,
		BizId:         bizIDText,
		BizType:       bizType,
		UserId:        userID,
		SymbolId:      symbolID,
		MarketType:    marketType,
		Source:        int64(trade.SourceType_SOURCE_TYPE_TASK),
		EventStatus:   int64(trade.EventStatus_EVENT_STATUS_PENDING),
		MaxRetryCount: 3,
		NextRetryAt:   now,
		Payload:       payload,
		CreateTimes:   now,
		UpdateTimes:   now,
	})
	return err
}
