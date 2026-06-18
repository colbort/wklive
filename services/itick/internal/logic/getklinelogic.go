package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetKlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetKlineLogic {
	return &GetKlineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取K线
func (l *GetKlineLogic) GetKline(in *itick.GetKlineReq) (*itick.GetKlineResp, error) {
	interval := l.buildInterval(in.KType)
	model := l.svcCtx.Factory.New(in.CategoryCode, interval)
	if model == nil {
		return &itick.GetKlineResp{}, nil
	}
	endTs := in.EndTs
	if endTs <= 0 {
		endTs = time.Now().UnixMilli() + 1
	}
	result, err := model.FindBeforeTsByMarketSymbol(l.ctx, in.Market, in.Symbol, endTs, in.Limit)
	if err != nil {
		return nil, err
	}

	if repaired, err := l.ensureKlineDataComplete(in, interval, endTs, in.Limit, len(result)); err != nil {
		l.Errorf("ensure kline data complete failed, category=%s market=%s symbol=%s interval=%s err=%v", in.CategoryCode, in.Market, in.Symbol, interval, err)
	} else if repaired {
		result, err = model.FindBeforeTsByMarketSymbol(l.ctx, in.Market, in.Symbol, endTs, in.Limit)
		if err != nil {
			return nil, err
		}
	}

	data := make([]*itick.Kline, 0)
	for _, item := range result {
		data = append(data, toKlineProto(in.KType, item))
	}
	return &itick.GetKlineResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}

func (l *GetKlineLogic) buildInterval(kType itick.KlineType) string {
	return utils.KlineTypeToInterval(kType)
}

func (l *GetKlineLogic) ensureKlineDataComplete(
	in *itick.GetKlineReq,
	interval string,
	endTs int64,
	limit int64,
	currentListLen int,
) (bool, error) {
	if interval == "" {
		return false, nil
	}
	if limit <= 0 {
		limit = 100
	}

	progress, err := l.svcCtx.ItickKlineSyncProgressModel.FindOrCreate(
		l.ctx,
		in.CategoryCode,
		in.Market,
		in.Symbol,
		interval,
	)
	if err != nil {
		return false, err
	}

	needRecent, needHistory := l.needRepairKline(progress, interval, endTs, limit, currentListLen)
	if !needRecent && !needHistory {
		return false, nil
	}

	lockKey := fmt.Sprintf("itick:get_kline_repair:%s:%s:%s:%s", in.CategoryCode, in.Market, in.Symbol, interval)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	distLock := utils.NewRedisLock(l.svcCtx.LockRedis)
	if err := distLock.Acquire(l.ctx, lockKey, lockValue, 30*time.Second); err != nil {
		if errors.Is(err, utils.ErrLockNotAcquired) {
			return l.waitKlineRepairDone(lockKey, in, interval, progress.UpdateTimes)
		}
		return false, err
	}
	defer func() {
		if err := distLock.Release(context.Background(), lockKey, lockValue); err != nil {
			l.Errorf("release get kline repair lock failed, key=%s err=%v", lockKey, err)
		}
	}()
	renewCtx, renewCancel := context.WithCancel(l.ctx)
	defer renewCancel()
	go l.renewKlineRepairLock(renewCtx, distLock, lockKey, lockValue, 10*time.Second, 30*time.Second)

	progress, err = l.svcCtx.ItickKlineSyncProgressModel.FindOneByCategoryCodeMarketSymbolIntervalNoCache(
		l.ctx,
		in.CategoryCode,
		in.Market,
		in.Symbol,
		interval,
	)
	if err != nil {
		return false, err
	}
	needRecent, needHistory = l.needRepairKline(progress, interval, endTs, limit, currentListLen)
	if !needRecent && !needHistory {
		return true, nil
	}

	now := time.Now().UnixMilli()
	mode := "on-demand"
	if err := l.svcCtx.ItickKlineSyncProgressModel.UpdateSyncStart(l.ctx, progress.Id, mode, now); err != nil {
		l.Errorf("update kline on-demand progress start failed, id=%d err=%v", progress.Id, err)
	}

	worker := NewSyncKlinesWorker(l.ctx, l.svcCtx, nil, "", "")
	job := KlineJob{
		ApiUrl:   l.svcCtx.Config.Itick.ApiUrl,
		Token:    l.svcCtx.Config.Itick.Token,
		Category: in.CategoryCode,
		Market:   in.Market,
		Symbol:   in.Symbol,
		KType:    int32(in.KType),
	}

	latestTs := progress.LatestTs
	contiguousTs := progress.ContiguousTs
	if contiguousTs == 0 && latestTs > 0 {
		contiguousTs = latestTs
	}
	recentCheckTs := progress.RecentCheckTs
	oldestTs := progress.OldestTs
	fullSynced := progress.FullSynced
	newCount := 0

	if needRecent {
		catchup, err := worker.syncCatchup(job, interval, contiguousTs, now)
		if err != nil {
			_ = l.svcCtx.ItickKlineSyncProgressModel.UpdateSyncFail(l.ctx, progress.Id, mode, time.Now().UnixMilli(), err.Error())
			return false, err
		}
		latestTs = maxInt64(latestTs, catchup.LatestTs)
		oldestTs = minNonZeroInt64(oldestTs, catchup.OldestTs)
		if catchup.ReachedBase && catchup.LatestTs > contiguousTs {
			contiguousTs = catchup.LatestTs
		}
		recentCheckTs = now
		newCount += catchup.NewCount
	}

	if needHistory && fullSynced == 0 {
		history, err := worker.syncHistory(job, interval, oldestTs, now)
		if err != nil {
			_ = l.svcCtx.ItickKlineSyncProgressModel.UpdateSyncFail(l.ctx, progress.Id, mode, time.Now().UnixMilli(), err.Error())
			return false, err
		}
		latestTs = maxInt64(latestTs, history.LatestTs)
		oldestTs = minNonZeroInt64(oldestTs, history.OldestTs)
		if history.FullSynced {
			fullSynced = 1
		}
		newCount += history.NewCount
	}

	now = time.Now().UnixMilli()
	msg := fmt.Sprintf("按需补齐成功，新增=%d", newCount)
	if err := l.svcCtx.ItickKlineSyncProgressModel.UpdateSyncSuccess(
		l.ctx,
		progress.Id,
		mode,
		latestTs,
		contiguousTs,
		recentCheckTs,
		oldestTs,
		fullSynced,
		now,
		msg,
	); err != nil {
		return false, err
	}

	return newCount > 0, nil
}

func (l *GetKlineLogic) needRepairKline(
	progress *models.TItickKlineSyncProgress,
	interval string,
	endTs int64,
	limit int64,
	currentListLen int,
) (bool, bool) {
	now := time.Now().UnixMilli()
	lastClosedTs := utils.LastClosedTs(now, interval)
	needRecent := lastClosedTs > 0 && endTs > lastClosedTs && progress.ContiguousTs < lastClosedTs

	intervalMs := utils.IntervalMillis(interval)
	wantOldestTs := int64(0)
	if intervalMs > 0 {
		wantOldestTs = endTs - intervalMs*(limit-1)
	}
	needHistory := progress.FullSynced == 0 &&
		(currentListLen < int(limit) || (wantOldestTs > 0 && (progress.OldestTs == 0 || wantOldestTs < progress.OldestTs)))

	return needRecent, needHistory
}

func (l *GetKlineLogic) waitKlineRepairDone(
	lockKey string,
	in *itick.GetKlineReq,
	interval string,
	lastUpdateTimes int64,
) (bool, error) {
	waitCtx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return false, nil
		case <-ticker.C:
			progress, err := l.svcCtx.ItickKlineSyncProgressModel.FindOneByCategoryCodeMarketSymbolIntervalNoCache(
				waitCtx,
				in.CategoryCode,
				in.Market,
				in.Symbol,
				interval,
			)
			if err == nil && progress.UpdateTimes > lastUpdateTimes {
				return true, nil
			}
			if err != nil && !errors.Is(err, models.ErrNotFound) {
				return false, err
			}

			exists, err := l.svcCtx.LockRedis.Exists(waitCtx, lockKey).Result()
			if err != nil {
				return false, err
			}
			if exists == 0 {
				return true, nil
			}
		}
	}
}

func (l *GetKlineLogic) renewKlineRepairLock(
	ctx context.Context,
	distLock *utils.RedisLock,
	lockKey string,
	lockValue string,
	interval time.Duration,
	ttl time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := distLock.Refresh(context.Background(), lockKey, lockValue, ttl); err != nil {
				l.Errorf("refresh get kline repair lock failed, key=%s err=%v", lockKey, err)
				return
			}
		}
	}
}

func maxInt64(a, b int64) int64 {
	if b > a {
		return b
	}
	return a
}

func minNonZeroInt64(a, b int64) int64 {
	if b <= 0 {
		return a
	}
	if a <= 0 || b < a {
		return b
	}
	return a
}
