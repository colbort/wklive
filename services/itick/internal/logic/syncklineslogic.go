package logic

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/time/rate"
)

type SyncKlinesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncKlinesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncKlinesLogic {
	return &SyncKlinesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步K线数据 （定时任务）
func (l *SyncKlinesLogic) SyncKlines(in *itick.SyncKlinesReq) (*itick.SyncKlinesResp, error) {
	if strings.TrimSpace(in.ApiUrl) == "" {
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(i18n.ApiURLRequired, i18n.Translate(i18n.ApiURLRequired, l.ctx)),
		}, nil
	}
	if strings.TrimSpace(in.ApiToken) == "" {
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(i18n.ApiTokenRequired, i18n.Translate(i18n.ApiTokenRequired, l.ctx)),
		}, nil
	}

	// 业务维度锁，避免相同源重复同步
	sum := md5.Sum([]byte(strings.TrimSpace(in.ApiUrl) + "|" + strings.TrimSpace(in.ApiToken)))
	lockKey := fmt.Sprintf("itick:sync_klines:%x", sum)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	distLock := utils.NewRedisLock(l.svcCtx.LockRedis)

	// 先抢锁，锁初始 30s，后续由 worker 自动续期
	if err := distLock.Acquire(l.ctx, lockKey, lockValue, 30*time.Second); err != nil {
		if errors.Is(err, utils.ErrLockNotAcquired) {
			return &itick.SyncKlinesResp{
				Base: helper.GetErrResp(i18n.SyncTaskAlreadyRunning, i18n.Translate(i18n.SyncTaskAlreadyRunning, l.ctx)),
			}, nil
		}

		logx.Errorf("acquire lock failed, key=%s err=%v", lockKey, err)
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(i18n.DistributedLockAcquireFailed, i18n.Translate(i18n.DistributedLockAcquireFailed, l.ctx)),
		}, nil
	}

	taskNo := fmt.Sprintf("sync_klines_%d", time.Now().UnixNano())
	now := cutils.NowMillis()

	_, err := l.svcCtx.ItickSyncTaskModel.Insert(l.ctx, &models.TItickSyncTask{
		TaskNo:      taskNo,
		TaskType:    "sync_klines",
		BizId:       0,
		Status:      0,
		Message:     "任务已提交",
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		_ = distLock.Release(l.ctx, lockKey, lockValue)

		logx.Errorf("create sync task failed, err=%v", err)
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(i18n.SyncTaskCreateFailed, i18n.Translate(i18n.SyncTaskCreateFailed, l.ctx)),
		}, nil
	}

	reqCopy := &itick.SyncKlinesReq{
		ApiUrl:   in.GetApiUrl(),
		ApiToken: in.GetApiToken(),
		WsUrl:    in.GetWsUrl(),
	}

	go func(taskNo string, reqCopy *itick.SyncKlinesReq, lockKey, lockValue string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 12*time.Hour)
		defer cancel()

		worker := NewSyncKlinesWorker(bgCtx, l.svcCtx, distLock, lockKey, lockValue)
		worker.Run(taskNo, reqCopy)
	}(taskNo, reqCopy, lockKey, lockValue)

	return &itick.SyncKlinesResp{
		Base: helper.OkResp(),
	}, nil
}

type SyncKlinesWorker struct {
	ctx          context.Context
	svcCtx       *svc.ServiceContext
	lock         *utils.RedisLock
	lockKey      string
	lockValue    string
	httpClient   *http.Client
	itickLimiter *rate.Limiter
	logx.Logger
}

func NewSyncKlinesWorker(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	lock *utils.RedisLock,
	lockKey string,
	lockValue string,
) *SyncKlinesWorker {
	return &SyncKlinesWorker{
		ctx:       ctx,
		svcCtx:    svcCtx,
		lock:      lock,
		lockKey:   lockKey,
		lockValue: lockValue,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		// 500次/分钟 = 500.0/60 次/秒
		// burst 这里给 1，最稳，不会突然打一波
		itickLimiter: rate.NewLimiter(rate.Limit(400.0/60.0), 1),
		Logger:       logx.WithContext(ctx),
	}
}

type KlineJob struct {
	ApiUrl   string
	ApiToken string
	Category string
	Market   string
	Symbol   string
	KType    int32
}

type klineSyncResult struct {
	LatestTs    int64
	OldestTs    int64
	NewCount    int
	ReachedBase bool
	FullSynced  bool
}

func (w *SyncKlinesWorker) Run(taskNo string, in *itick.SyncKlinesReq) {
	// 自动续期协程
	renewCtx, renewCancel := context.WithCancel(w.ctx)
	defer renewCancel()

	go w.autoRenewLock(renewCtx, 10*time.Second, 30*time.Second)

	defer func() {
		if err := w.lock.Release(context.Background(), w.lockKey, w.lockValue); err != nil {
			w.Errorf("release lock failed, key=%s err=%v", w.lockKey, err)
		}

		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("同步任务 panic: %v", r)
			w.Errorf(errMsg)
			_ = w.updateTaskStatus(taskNo, 3, errMsg)
		}
	}()

	_ = w.updateTaskStatus(taskNo, 1, "同步中")

	if err := w.doSync(in); err != nil {
		w.Errorf("sync klines failed, taskNo=%s err=%v", taskNo, err)
		_ = w.updateTaskStatus(taskNo, 3, err.Error())
		return
	}

	_ = w.updateTaskStatus(taskNo, 2, "同步成功")
}

func (w *SyncKlinesWorker) autoRenewLock(ctx context.Context, interval, ttl time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := w.lock.Refresh(context.Background(), w.lockKey, w.lockValue, ttl)
			if err != nil {
				w.Errorf("refresh lock failed, key=%s err=%v", w.lockKey, err)
				return
			}
		}
	}
}

func (w *SyncKlinesWorker) doSync(in *itick.SyncKlinesReq) error {
	const workerCount = 8

	jobs := make(chan KlineJob, 1000)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if err := w.syncOneJob(job); err != nil {
					w.Errorf("sync job failed, category=%s market=%s symbol=%s kType=%d err=%v",
						job.Category, job.Market, job.Symbol, job.KType, err)
				}
			}
		}()
	}

	cursor := int64(0)
	pageSize := int64(500)

	for {
		products, _, err := w.svcCtx.ItickProductModel.FindPage(
			w.ctx,
			0,
			"", "", "",
			0, 0, "",
			cursor,
			pageSize,
		)
		if err != nil {
			close(jobs)
			wg.Wait()
			return i18n.StatusError(w.ctx, i18n.InternalServerError)
		}
		if len(products) == 0 {
			break
		}

		for _, product := range products {
			category := utils.NormalizeCategory(product.CategoryCode)
			market := utils.NormalizeMarket(product.Market)
			symbol := utils.NormalizeSymbol(product.Symbol)

			if category == "" || market == "" || symbol == "" {
				continue
			}
			if !utils.IsSupportedKlineCategory(category) {
				continue
			}

			for _, kType := range utils.DefaultKTypes {
				jobs <- KlineJob{
					ApiUrl:   in.ApiUrl,
					ApiToken: in.ApiToken,
					Category: category,
					Market:   market,
					Symbol:   symbol,
					KType:    kType,
				}
			}
		}

		if len(products) < int(pageSize) {
			break
		}
		cursor = products[len(products)-1].Id
	}

	close(jobs)
	wg.Wait()
	return nil
}

func (w *SyncKlinesWorker) syncOneJob(job KlineJob) error {
	interval := utils.KTypeToIntervalName(job.KType)
	if interval == "" {
		return i18n.StatusError(w.ctx, i18n.ParamError)
	}

	progress, err := w.svcCtx.ItickKlineSyncProgressModel.FindOrCreate(
		w.ctx,
		job.Category,
		job.Market,
		job.Symbol,
		interval,
	)
	if err != nil {
		return i18n.StatusError(w.ctx, i18n.InternalServerError)
	}

	now := cutils.NowMillis()
	mode := "mixed"

	if err := w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncStart(w.ctx, progress.Id, mode, now); err != nil {
		w.Errorf("update progress start failed, id=%d err=%v", progress.Id, err)
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

	catchup, err := w.syncCatchup(job, interval, contiguousTs, now)
	if err != nil {
		_ = w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncFail(w.ctx, progress.Id, mode, cutils.NowMillis(), err.Error())
		return err
	}
	if catchup.LatestTs > latestTs {
		latestTs = catchup.LatestTs
	}
	if oldestTs == 0 || (catchup.OldestTs > 0 && catchup.OldestTs < oldestTs) {
		oldestTs = catchup.OldestTs
	}
	newCount += catchup.NewCount
	if catchup.ReachedBase && catchup.LatestTs > contiguousTs {
		contiguousTs = catchup.LatestTs
	}

	recent, err := w.syncRecentCheck(job, interval, now)
	if err != nil {
		_ = w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncFail(w.ctx, progress.Id, mode, cutils.NowMillis(), err.Error())
		return err
	}
	if recent.LatestTs > latestTs {
		latestTs = recent.LatestTs
	}
	if oldestTs == 0 || (recent.OldestTs > 0 && recent.OldestTs < oldestTs) {
		oldestTs = recent.OldestTs
	}
	if contiguousTs == 0 && recent.LatestTs > 0 {
		contiguousTs = recent.LatestTs
	}
	recentCheckTs = now
	newCount += recent.NewCount

	if fullSynced == 0 {
		history, err := w.syncHistory(job, interval, oldestTs, now)
		if err != nil {
			_ = w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncFail(w.ctx, progress.Id, mode, cutils.NowMillis(), err.Error())
			return err
		}
		if history.LatestTs > latestTs {
			latestTs = history.LatestTs
		}
		if oldestTs == 0 || (history.OldestTs > 0 && history.OldestTs < oldestTs) {
			oldestTs = history.OldestTs
		}
		if history.FullSynced {
			fullSynced = 1
		}
		newCount += history.NewCount
	}

	msg := fmt.Sprintf("同步成功，模式=%s，新增=%d", mode, newCount)
	return w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncSuccess(
		w.ctx,
		progress.Id,
		mode,
		latestTs,
		contiguousTs,
		recentCheckTs,
		oldestTs,
		fullSynced,
		cutils.NowMillis(),
		msg,
	)
}

func (w *SyncKlinesWorker) syncCatchup(job KlineJob, interval string, contiguousTs int64, now int64) (klineSyncResult, error) {
	const limit = 100

	to := utils.LastClosedTs(now, interval)
	if to <= 0 {
		return klineSyncResult{}, nil
	}
	if contiguousTs >= to {
		return klineSyncResult{LatestTs: contiguousTs, ReachedBase: true}, nil
	}

	maxPages := 1
	intervalMs := utils.IntervalMillis(interval)
	if contiguousTs > 0 && intervalMs > 0 {
		gapBars := (to - contiguousTs) / intervalMs
		maxPages = int(gapBars/int64(limit)) + 2
		if maxPages < 1 {
			maxPages = 1
		}
		if maxPages > 200 {
			maxPages = 200
		}
	}

	result, err := w.syncBackwardRange(job, interval, to+1, contiguousTs, to, limit, maxPages)
	if err != nil {
		return klineSyncResult{}, err
	}
	if contiguousTs == 0 {
		result.ReachedBase = true
	}
	if result.ReachedBase && result.LatestTs < to {
		result.LatestTs = to
	}
	return result, nil
}

func (w *SyncKlinesWorker) syncRecentCheck(job KlineJob, interval string, now int64) (klineSyncResult, error) {
	limit := utils.RecentCheckBars(interval)
	if limit <= 0 {
		limit = 3
	}
	to := utils.LastClosedTs(now, interval)
	if to <= 0 {
		to = now
	}
	return w.syncBackwardRange(job, interval, to+1, 0, to, limit, 1)
}

func (w *SyncKlinesWorker) syncHistory(job KlineJob, interval string, oldestTs int64, now int64) (klineSyncResult, error) {
	const limit = 300

	et := now
	if oldestTs > 0 {
		et = oldestTs - 1
	}

	result, err := w.syncBackwardRange(job, interval, et, 0, et, limit, 3)
	if err != nil {
		return klineSyncResult{}, err
	}
	if result.NewCount == 0 || result.NewCount < limit {
		result.FullSynced = true
	}
	return result, nil
}

func (w *SyncKlinesWorker) syncBackwardRange(job KlineJob, interval string, et int64, stopAtTs int64, maxAcceptTs int64, limit int, maxPages int) (klineSyncResult, error) {
	var result klineSyncResult

	for page := 0; page < maxPages; page++ {
		resp, err := w.getKlineFromItick(w.ctx, job, et, limit)
		if err != nil {
			return result, err
		}
		if len(resp.Data) == 0 {
			result.ReachedBase = true
			return result, nil
		}

		list := make([]*models.CoinKline, 0, len(resp.Data))
		minTs := resp.Data[0].T

		for _, item := range resp.Data {
			if item.T < minTs {
				minTs = item.T
			}
			if stopAtTs > 0 && item.T <= stopAtTs {
				result.ReachedBase = true
				continue
			}
			if maxAcceptTs > 0 && item.T > maxAcceptTs {
				continue
			}

			list = append(list, w.toCoinKline(job, interval, item))
			if item.T > result.LatestTs {
				result.LatestTs = item.T
			}
			if result.OldestTs == 0 || item.T < result.OldestTs {
				result.OldestTs = item.T
			}
		}

		if len(list) > 0 {
			if err := w.bulkUpsertKlines(job.Category, interval, list); err != nil {
				return result, err
			}
			result.NewCount += len(list)
		}

		if result.ReachedBase || len(resp.Data) < limit || minTs <= 0 || minTs >= et {
			return result, nil
		}
		et = minTs - 1
	}

	return result, nil
}

func (w *SyncKlinesWorker) toCoinKline(job KlineJob, interval string, item ItickKlineItem) *models.CoinKline {
	return &models.CoinKline{
		CategoryCode: job.Category,
		Market:       job.Market,
		Symbol:       job.Symbol,
		Interval:     interval,
		Ts:           item.T,
		Open:         item.O,
		High:         item.H,
		Low:          item.L,
		Close:        item.C,
		Volume:       item.V,
		Turnover:     item.Tu,
	}
}

func (w *SyncKlinesWorker) bulkUpsertKlines(category, interval string, list []*models.CoinKline) error {
	if len(list) == 0 {
		return nil
	}

	model := w.svcCtx.Factory.New(category, interval)
	if model == nil {
		return fmt.Errorf("invalid kline model, category=%s interval=%s", category, interval)
	}

	return model.BulkUpsertBySymbolTs(w.ctx, list)
}

func (w *SyncKlinesWorker) getIncrementalPages(kType int32) int {
	switch kType {
	case 1: // 1m
		return 5
	case 2: // 5m
		return 2
	default:
		return 1
	}
}

type ItickKlineResponse struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data []ItickKlineItem `json:"data"`
}

type ItickKlineItem struct {
	T  int64   `json:"t"`
	O  float64 `json:"o"`
	H  float64 `json:"h"`
	L  float64 `json:"l"`
	C  float64 `json:"c"`
	V  float64 `json:"v"`
	Tu float64 `json:"tu"`
}

func (w *SyncKlinesWorker) getKlineFromItick(
	ctx context.Context,
	job KlineJob,
	et int64,
	limit int,
) (*ItickKlineResponse, error) {
	apiURL := strings.TrimSpace(job.ApiUrl)
	token := strings.TrimSpace(job.ApiToken)
	category := strings.ToLower(strings.TrimSpace(job.Category))
	market := strings.ToUpper(strings.TrimSpace(job.Market))
	symbol := strings.TrimSpace(job.Symbol)

	if apiURL == "" {
		return nil, i18n.StatusError(ctx, i18n.APIURLIsRequired)
	}
	if token == "" {
		return nil, i18n.StatusError(ctx, i18n.TokenRequired)
	}
	if category == "" {
		return nil, i18n.StatusError(ctx, i18n.CategoryRequired)
	}
	if market == "" {
		return nil, i18n.StatusError(ctx, i18n.MarketRequired)
	}
	if symbol == "" {
		return nil, i18n.StatusError(ctx, i18n.SymbolRequired)
	}

	base, err := url.Parse(apiURL)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.ParamError)
	}

	base.Path = path.Join(strings.TrimRight(base.Path, "/"), fmt.Sprintf("/%s/kline", category))

	q := base.Query()
	q.Set("region", market)
	q.Set("code", symbol)
	q.Set("kType", fmt.Sprintf("%d", job.KType))
	q.Set("limit", fmt.Sprintf("%d", limit))
	if et > 0 {
		q.Set("et", fmt.Sprintf("%d", et))
	}
	base.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.InternalServerError)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("token", token)

	// 在真正发请求前限流
	if w.itickLimiter != nil {
		if err := w.itickLimiter.Wait(ctx); err != nil {
			return nil, i18n.StatusError(ctx, i18n.ServiceUnavailable)
		}
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.ServiceUnavailable)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		logx.Errorf("itick returned non-200 status: %d body=%s", resp.StatusCode, string(raw))
		return nil, i18n.StatusError(ctx, i18n.ServiceUnavailable)
	}

	var out ItickKlineResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, i18n.StatusError(ctx, i18n.InternalServerError)
	}
	if out.Code != 0 {
		return &out, i18n.StatusError(ctx, i18n.ServiceUnavailable)
	}

	return &out, nil
}

func (w *SyncKlinesWorker) updateTaskStatus(taskNo string, status int64, message string) error {
	return w.svcCtx.ItickSyncTaskModel.UpdateStatusByTaskNo(
		w.ctx,
		taskNo,
		status,
		message,
		cutils.NowMillis(),
	)
}
