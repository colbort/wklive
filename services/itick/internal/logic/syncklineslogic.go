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
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ApiURLRequired, l.ctx)),
		}, nil
	}
	if strings.TrimSpace(in.ApiToken) == "" {
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ApiTokenRequired, l.ctx)),
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
				Base: helper.GetErrResp(1, i18n.Translate(i18n.SyncTaskAlreadyRunning, l.ctx)),
			}, nil
		}

		logx.Errorf("acquire lock failed, key=%s err=%v", lockKey, err)
		return &itick.SyncKlinesResp{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.DistributedLockAcquireFailed, l.ctx)),
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
			Base: helper.GetErrResp(1, i18n.Translate(i18n.SyncTaskCreateFailed, l.ctx)),
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
	Category string
	Market   string
	Symbol   string
	KType    int32
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
				if err := w.syncOneJob(in.ApiUrl, in.ApiToken, job); err != nil {
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
			0, 0,
			cursor,
			pageSize,
		)
		if err != nil {
			close(jobs)
			wg.Wait()
			return fmt.Errorf("find products failed: %w", err)
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

func (w *SyncKlinesWorker) syncOneJob(apiURL, token string, job KlineJob) error {
	interval := utils.KTypeToIntervalName(job.KType)
	if interval == "" {
		return fmt.Errorf("unknown kType: %d", job.KType)
	}

	progress, err := w.svcCtx.ItickKlineSyncProgressModel.FindOrCreate(
		w.ctx,
		job.Category,
		job.Market,
		job.Symbol,
		interval,
	)
	if err != nil {
		return fmt.Errorf("find or create progress failed: %w", err)
	}

	now := cutils.NowMillis()

	mode := "incremental"
	et := now
	maxPages := w.getIncrementalPages(job.KType)

	// 历史没补完：从 oldest_ts 往前补
	if progress.FullSynced == 0 {
		mode = "history"
		maxPages = 3

		if progress.OldestTs > 0 {
			et = progress.OldestTs - 1
		} else {
			et = now
		}
	}

	if err := w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncStart(w.ctx, progress.Id, mode, now); err != nil {
		w.Errorf("update progress start failed, id=%d err=%v", progress.Id, err)
	}

	const limit = 300
	latestTs := progress.LatestTs
	oldestTs := progress.OldestTs
	fullSynced := progress.FullSynced
	newCount := 0

	for page := 0; page < maxPages; page++ {
		resp, err := w.getKlineFromItick(
			w.ctx,
			apiURL,
			token,
			job.Category,
			job.Market,
			job.Symbol,
			int(job.KType),
			et,
			limit,
		)
		if err != nil {
			_ = w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncFail(
				w.ctx,
				progress.Id,
				mode,
				cutils.NowMillis(),
				err.Error(),
			)
			return err
		}

		if len(resp.Data) == 0 {
			if mode == "history" {
				fullSynced = 1
			}
			break
		}

		minTs := resp.Data[0].T
		maxTs := resp.Data[0].T

		for _, item := range resp.Data {
			if item.T < minTs {
				minTs = item.T
			}
			if item.T > maxTs {
				maxTs = item.T
			}

			// 增量模式：只处理比 latest_ts 新的数据
			if mode == "incremental" && item.T <= latestTs {
				continue
			}

			data := &models.CoinKline{
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

			if err := w.svcCtx.Writer.Enqueue(data); err != nil {
				w.Errorf("enqueue kline error: category=%s market=%s symbol=%s interval=%s ts=%d err=%v",
					data.CategoryCode, data.Market, data.Symbol, data.Interval, data.Ts, err)
				continue
			}

			newCount++

			if latestTs == 0 || item.T > latestTs {
				latestTs = item.T
			}
			if oldestTs == 0 || item.T < oldestTs {
				oldestTs = item.T
			}
		}

		if mode == "incremental" {
			// 最近几页里，如果已经翻到 latest_ts 之前，就停
			if minTs <= progress.LatestTs {
				break
			}
			et = minTs - 1
			if et <= 0 {
				break
			}
			continue
		}

		// history 模式
		if len(resp.Data) < limit {
			fullSynced = 1
			break
		}
		if minTs <= 0 || minTs >= et {
			break
		}

		et = minTs - 1
	}

	msg := fmt.Sprintf("同步成功，模式=%s，新增=%d", mode, newCount)
	return w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncSuccess(
		w.ctx,
		progress.Id,
		mode,
		latestTs,
		oldestTs,
		fullSynced,
		cutils.NowMillis(),
		msg,
	)
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
	apiURL, token, category, market, symbol string,
	kType int,
	et int64,
	limit int,
) (*ItickKlineResponse, error) {
	apiURL = strings.TrimSpace(apiURL)
	token = strings.TrimSpace(token)
	category = strings.ToLower(strings.TrimSpace(category))
	market = strings.ToUpper(strings.TrimSpace(market))
	symbol = strings.TrimSpace(symbol)

	if apiURL == "" {
		return nil, errors.New(i18n.Translate(i18n.APIURLIsRequired, ctx))
	}
	if token == "" {
		return nil, errors.New(i18n.Translate(i18n.TokenRequired, ctx))
	}
	if category == "" {
		return nil, errors.New(i18n.Translate(i18n.CategoryRequired, ctx))
	}
	if market == "" {
		return nil, errors.New(i18n.Translate(i18n.MarketRequired, ctx))
	}
	if symbol == "" {
		return nil, errors.New(i18n.Translate(i18n.SymbolRequired, ctx))
	}

	base, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid apiURL: %w", err)
	}

	base.Path = path.Join(strings.TrimRight(base.Path, "/"), fmt.Sprintf("/%s/kline", category))

	q := base.Query()
	q.Set("region", market)
	q.Set("code", symbol)
	q.Set("kType", fmt.Sprintf("%d", kType))
	q.Set("limit", fmt.Sprintf("%d", limit))
	if et > 0 {
		q.Set("et", fmt.Sprintf("%d", et))
	}
	base.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("token", token)

	// 在真正发请求前限流
	if w.itickLimiter != nil {
		if err := w.itickLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("itick rate limit wait failed: %w", err)
		}
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request itick kline failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("itick returned non-200 status: %d body=%s", resp.StatusCode, string(raw))
	}

	var out ItickKlineResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode kline response failed: %w", err)
	}
	if out.Code != 0 {
		return &out, fmt.Errorf("itick business error: code=%d msg=%s", out.Code, out.Msg)
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
