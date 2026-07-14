package kline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/services/itick/internal/pkg/utils"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/time/rate"
)

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
	Token    string
	Category string
	Market   string
	Symbol   string
	KType    int32
}

type SyncResult struct {
	NewCount int
}

// FetchProductHistory walks backwards until iTick has no older data and
// persists every returned page to MongoDB. Page size is internal and is not
// exposed by the Admin API.
func (w *SyncKlinesWorker) FetchProductHistory(job KlineJob, interval string, endTs int64) (SyncResult, error) {
	const pageSize = 500
	if endTs <= 0 {
		endTs = time.Now().UnixMilli() + 1
	}
	return w.syncBackwardRange(job, interval, endTs, endTs, pageSize)
}

const (
	activeProductsKey   = "itick:v1:active_products"
	reconcileBatchSize  = 10
	reconcileKType      = int32(1)
	reconcileWindowBars = 30
)

type reconcileGroup struct {
	category string
	market   string
	exchange string
	products []*models.TItickProduct
}

// RunReconcile performs only the bounded recent-window correction used by the
// five-minute scheduler. It never walks backwards through historical pages.
func (w *SyncKlinesWorker) RunReconcile(taskNo, apiURL, token string) {
	w.runTask(taskNo, "五分钟校准中", func() error {
		return w.doReconcile(apiURL, token)
	})
}

func (w *SyncKlinesWorker) runTask(taskNo, runningMessage string, run func() error) {
	renewCtx, renewCancel := context.WithCancel(w.ctx)
	defer renewCancel()
	if w.lock != nil {
		go w.autoRenewLock(renewCtx, 10*time.Second, 30*time.Second)
	}
	defer func() {
		if w.lock != nil {
			if err := w.lock.Release(context.Background(), w.lockKey, w.lockValue); err != nil {
				w.Errorf("release lock failed, key=%s err=%v", w.lockKey, err)
			}
		}
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("同步任务 panic: %v", r)
			w.Errorf("%s", errMsg)
			_ = w.updateTaskStatus(taskNo, 3, errMsg)
		}
	}()
	_ = w.updateTaskStatus(taskNo, 1, runningMessage)
	if err := run(); err != nil {
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

func (w *SyncKlinesWorker) doReconcile(apiURL, token string) error {
	products, err := w.loadActiveProducts()
	if err != nil {
		return err
	}
	groups := make(map[string]*reconcileGroup)
	for _, product := range products {
		category := utils.NormalizeCategory(product.CategoryCode)
		market := utils.NormalizeMarket(product.Market)
		symbol := utils.NormalizeSymbol(product.Symbol)
		if category == "" || market == "" || symbol == "" || !utils.IsSupportedKlineCategory(category) {
			continue
		}
		product.CategoryCode, product.Market, product.Symbol = category, market, symbol
		key := category + "\x00" + market + "\x00" + strings.TrimSpace(product.Exchange)
		if groups[key] == nil {
			groups[key] = &reconcileGroup{category: category, market: market, exchange: strings.TrimSpace(product.Exchange)}
		}
		groups[key].products = append(groups[key].products, product)
	}

	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var failures []string
	for _, key := range keys {
		group := groups[key]
		for start := 0; start < len(group.products); start += reconcileBatchSize {
			end := min(start+reconcileBatchSize, len(group.products))
			if err := w.reconcileBatch(apiURL, token, group, group.products[start:end]); err != nil {
				failures = append(failures, err.Error())
			}
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("kline reconcile partial failure (%d): %s", len(failures), strings.Join(failures, "; "))
	}
	return nil
}

func (w *SyncKlinesWorker) loadActiveProducts() ([]*models.TItickProduct, error) {
	if w.svcCtx.BusRedis != nil {
		members, err := w.svcCtx.BusRedis.SMembers(w.ctx, activeProductsKey).Result()
		if err == nil && len(members) > 0 {
			ids := make([]int64, 0, len(members))
			for _, member := range members {
				id, parseErr := strconv.ParseInt(member, 10, 64)
				if parseErr == nil && id > 0 {
					ids = append(ids, id)
				}
			}
			if len(ids) > 0 {
				return w.svcCtx.ItickProductModel.FindByIds(w.ctx, ids)
			}
		}
		if err != nil {
			w.Errorf("load active products from redis failed, fallback to mysql: %v", err)
		}
	}

	var result []*models.TItickProduct
	var cursor int64
	for {
		page, err := w.svcCtx.ItickProductModel.FindActivePage(w.ctx, cursor, 500)
		if err != nil {
			return nil, err
		}
		result = append(result, page...)
		if len(page) < 500 {
			break
		}
		cursor = page[len(page)-1].Id
	}
	return result, nil
}

func (w *SyncKlinesWorker) reconcileBatch(apiURL, token string, group *reconcileGroup, products []*models.TItickProduct) error {
	codes := make([]string, 0, len(products))
	for _, product := range products {
		codes = append(codes, product.Symbol)
	}
	data, err := w.getBatchKlines(w.ctx, apiURL, token, group.category, group.market, group.exchange, codes, reconcileKType, reconcileWindowBars)
	if err != nil {
		return fmt.Errorf("category=%s market=%s codes=%s: %w", group.category, group.market, strings.Join(codes, ","), err)
	}
	interval := utils.KTypeToIntervalName(reconcileKType)
	now := cutils.NowMillis()
	lastClosed := utils.LastClosedTs(now, interval)
	var failures []string
	for _, product := range products {
		if err := w.reconcileProduct(group, product, data, interval, now, lastClosed); err != nil {
			failures = append(failures, fmt.Sprintf("symbol=%s: %v", product.Symbol, err))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("batch product failures (%d): %s", len(failures), strings.Join(failures, "; "))
	}
	return nil
}

func (w *SyncKlinesWorker) reconcileProduct(group *reconcileGroup, product *models.TItickProduct,
	data map[string][]ItickKlineItem, interval string, now, lastClosed int64) error {
	items := data[product.Symbol]
	if items == nil {
		for symbol, candidate := range data {
			if strings.EqualFold(symbol, product.Symbol) {
				items = candidate
				break
			}
		}
	}
	if len(items) == 0 {
		return fmt.Errorf("batch response missing kline data")
	}
	list := make([]*models.CoinKline, 0, len(items))
	var latestTs, oldestTs int64
	job := KlineJob{Category: group.category, Market: group.market, Symbol: product.Symbol, KType: reconcileKType}
	for _, item := range items {
		if !validClosedKline(item, lastClosed, utils.IntervalMillis(interval)) {
			continue
		}
		list = append(list, w.toCoinKline(job, interval, item))
		if item.T > latestTs {
			latestTs = item.T
		}
		if oldestTs == 0 || item.T < oldestTs {
			oldestTs = item.T
		}
	}
	if err := w.bulkUpsertKlines(group.category, interval, list); err != nil {
		return err
	}
	if w.svcCtx.RebuildDerivedKlines != nil {
		if err := w.svcCtx.RebuildDerivedKlines(list); err != nil {
			return fmt.Errorf("rebuild derived klines: %w", err)
		}
	}
	progress, err := w.svcCtx.ItickKlineSyncProgressModel.FindOrCreate(w.ctx, group.category, group.market, product.Symbol, interval)
	if err != nil {
		return err
	}
	if latestTs == 0 {
		latestTs = progress.LatestTs
	}
	if oldestTs == 0 || (progress.OldestTs > 0 && progress.OldestTs < oldestTs) {
		oldestTs = progress.OldestTs
	}
	if err := w.svcCtx.ItickKlineSyncProgressModel.UpdateSyncSuccess(w.ctx, progress.Id, "reconcile", latestTs,
		progress.ContiguousTs, now, oldestTs, progress.FullSynced, now,
		fmt.Sprintf("五分钟校准成功，覆盖=%d", len(list))); err != nil {
		return err
	}
	return nil
}

func validClosedKline(item ItickKlineItem, lastClosed, intervalMs int64) bool {
	return item.T > 0 && item.T <= lastClosed && intervalMs > 0 && item.T%intervalMs == 0 &&
		item.L <= item.O && item.L <= item.C && item.H >= item.O && item.H >= item.C &&
		item.H >= item.L && item.V >= 0 && item.Tu >= 0
}

func (w *SyncKlinesWorker) getBatchKlines(ctx context.Context, apiURL, token, category, market, exchange string,
	codes []string, kType int32, limit int) (map[string][]ItickKlineItem, error) {
	base, err := url.Parse(strings.TrimSpace(apiURL))
	if err != nil {
		return nil, err
	}
	base.Path = path.Join(strings.TrimRight(base.Path, "/"), category, "klines")
	q := base.Query()
	q.Set("region", market)
	q.Set("codes", strings.Join(codes, ","))
	q.Set("kType", strconv.FormatInt(int64(kType), 10))
	q.Set("limit", strconv.Itoa(limit))
	if exchange != "" {
		q.Set("exchange", exchange)
	}
	base.RawQuery = q.Encode()
	if err := w.itickLimiter.Wait(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("token", strings.TrimSpace(token))
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("REST returned status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out struct {
		Code int                         `json:"code"`
		Msg  string                      `json:"msg"`
		Data map[string][]ItickKlineItem `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("REST rejected: code=%d msg=%s", out.Code, out.Msg)
	}
	return out.Data, nil
}

func (w *SyncKlinesWorker) syncBackwardRange(job KlineJob, interval string, et int64, maxAcceptTs int64, limit int) (SyncResult, error) {
	var result SyncResult
	for {
		resp, err := w.getSingleKline(w.ctx, job, et, limit)
		if err != nil {
			return result, err
		}
		if len(resp.Data) == 0 {
			return result, nil
		}

		list := make([]*models.CoinKline, 0, len(resp.Data))
		minTs := resp.Data[0].T

		for _, item := range resp.Data {
			if item.T < minTs {
				minTs = item.T
			}
			if maxAcceptTs > 0 && item.T > maxAcceptTs {
				continue
			}

			list = append(list, w.toCoinKline(job, interval, item))
		}

		if len(list) > 0 {
			if err := w.bulkUpsertKlines(job.Category, interval, list); err != nil {
				return result, err
			}
			if interval == "1m" && w.svcCtx.RebuildDerivedKlines != nil {
				if err := w.svcCtx.RebuildDerivedKlines(list); err != nil {
					return result, fmt.Errorf("rebuild derived klines: %w", err)
				}
			}
			result.NewCount += len(list)
		}

		if len(resp.Data) < limit || minTs <= 0 || minTs >= et {
			return result, nil
		}
		et = minTs - 1
	}
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

// getSingleKline is reserved for one-product historical pagination. The batch
// endpoint intentionally remains a latest-window API for reconciliation.
func (w *SyncKlinesWorker) getSingleKline(
	ctx context.Context,
	job KlineJob,
	et int64,
	limit int,
) (*ItickKlineResponse, error) {
	apiURL := strings.TrimSpace(job.ApiUrl)
	token := strings.TrimSpace(job.Token)
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
	base.Path = path.Join(strings.TrimRight(base.Path, "/"), category, "kline")
	q := base.Query()
	q.Set("region", market)
	q.Set("code", symbol)
	q.Set("kType", strconv.FormatInt(int64(job.KType), 10))
	q.Set("limit", strconv.Itoa(limit))
	if et > 0 {
		q.Set("et", strconv.FormatInt(et, 10))
	}
	base.RawQuery = q.Encode()
	if w.itickLimiter != nil {
		if err := w.itickLimiter.Wait(ctx); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("token", token)
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("REST returned status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out ItickKlineResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return &out, fmt.Errorf("REST rejected: code=%d msg=%s", out.Code, out.Msg)
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
