package kline

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

const (
	gapQueueKey      = "market:v1:kline:repair:queue"
	gapJobHashKey    = "market:v1:kline:repair:jobs"
	gapDeadHashKey   = "market:v1:kline:repair:dead"
	gapDoneKeyPrefix = "market:v2:kline:repair:done:"
	gapScanPageSize  = int64(2000)
	gapMaxAttempts   = 5
	minuteMs         = int64(time.Minute / time.Millisecond)
)

var claimGapJobScript = redis.NewScript(`
local ids = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1], 'LIMIT', 0, 1)
if #ids == 0 then return nil end
if redis.call('ZREM', KEYS[1], ids[1]) == 0 then return nil end
return {ids[1], redis.call('HGET', KEYS[2], ids[1])}
`)

type gapScanState struct {
	BeforeTs int64 `json:"beforeTs"`
	NewerTs  int64 `json:"newerTs"`
}

type GapRepairJob struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Market   string `json:"market"`
	Exchange string `json:"exchange,omitempty"`
	Symbol   string `json:"symbol"`
	StartTs  int64  `json:"startTs"`
	EndTs    int64  `json:"endTs"`
	Attempts int    `json:"attempts"`
}

type gapDeadLetter struct {
	Job      *GapRepairJob `json:"job"`
	Error    string        `json:"error"`
	FailedAt int64         `json:"failedAt"`
}

// GapRepairService incrementally walks old 1m history and stores repair jobs
// in Redis. Both scan cursors and jobs survive process restarts.
type GapRepairService struct {
	ctx          context.Context
	cancel       context.CancelFunc
	svcCtx       *svc.ServiceContext
	worker       *SyncKlinesWorker
	scanInterval time.Duration
	batchSize    int
	wg           sync.WaitGroup
}

func NewGapRepairService(parent context.Context, svcCtx *svc.ServiceContext, scanInterval time.Duration, batchSize int) *GapRepairService {
	if scanInterval <= 0 {
		scanInterval = time.Hour
	}
	if batchSize <= 0 {
		batchSize = 10
	}
	ctx, cancel := context.WithCancel(parent)
	worker := NewSyncKlinesWorker(ctx, svcCtx, nil, "", "")
	return &GapRepairService{ctx: ctx, cancel: cancel, svcCtx: svcCtx, worker: worker, scanInterval: scanInterval, batchSize: batchSize}
}

func (s *GapRepairService) Start(apiURL, token string) {
	s.wg.Add(2)
	go s.scanLoop()
	go s.repairLoop(apiURL, token)
}

func (s *GapRepairService) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *GapRepairService) scanLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()
	for {
		if err := s.scanOnce(); err != nil {
			logx.Errorf("scan historical kline gaps failed: %v", err)
		}
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *GapRepairService) scanOnce() error {
	products, err := s.worker.loadActiveProducts()
	if err != nil {
		return err
	}
	// Scan every closed minute. The five-minute reconciliation remains useful
	// for correction, but gap detection must not depend on an external scheduler.
	cutoff := time.Now().UnixMilli()/minuteMs*minuteMs - minuteMs
	return mr.MapReduceVoid(func(source chan<- *models.TMarketProduct) {
		for _, product := range products {
			source <- product
		}
	}, func(product *models.TMarketProduct, writer mr.Writer[error], _ func(error)) {
		if err := s.scanProduct(product, cutoff); err != nil {
			writer.Write(fmt.Errorf("product=%d symbol=%s: %w", product.Id, product.Symbol, err))
		}
	}, func(pipe <-chan error, _ func(error)) {
		for err := range pipe {
			logx.Errorf("scan product kline gaps failed, %v", err)
		}
	}, mr.WithContext(s.ctx), mr.WithWorkers(4))
}

func (s *GapRepairService) scanProduct(product *models.TMarketProduct, cutoff int64) error {
	category := strings.ToLower(strings.TrimSpace(product.CategoryCode))
	market := strings.ToUpper(strings.TrimSpace(product.Market))
	symbol := strings.ToUpper(strings.TrimSpace(product.Symbol))
	model := s.svcCtx.Factory.New(category, "1m")
	if model == nil {
		return nil
	}
	stateKey := fmt.Sprintf("market:v1:kline:gap_scan:%d", product.Id)
	var state gapScanState
	if raw, err := s.svcCtx.DataCache.Get(s.ctx, stateKey).Bytes(); err == nil {
		_ = json.Unmarshal(raw, &state)
	}
	var list []*models.CoinKline
	var err error
	if state.BeforeTs > 0 {
		list, err = model.FindBeforeTsByMarketSymbol(s.ctx, market, symbol, state.BeforeTs, gapScanPageSize)
	} else {
		list, err = model.FindLatestByMarketSymbol(s.ctx, market, symbol, gapScanPageSize)
	}
	if err != nil || len(list) == 0 {
		if err == nil && state.BeforeTs > 0 {
			return s.svcCtx.DataCache.Del(s.ctx, stateKey).Err()
		}
		return err
	}
	newerTs := state.NewerTs
	for _, bar := range list {
		if bar == nil || bar.Ts <= 0 {
			continue
		}
		if newerTs > 0 && newerTs-bar.Ts > minuteMs {
			end := min(newerTs-minuteMs, cutoff)
			if end >= bar.Ts+minuteMs {
				for _, job := range s.expectedGapJobs(product, bar.Ts+minuteMs, end) {
					if err := s.enqueue(job); err != nil {
						return err
					}
				}
			}
		}
		newerTs = bar.Ts
	}
	if len(list) < int(gapScanPageSize) {
		return s.svcCtx.DataCache.Del(s.ctx, stateKey).Err()
	}
	state = gapScanState{BeforeTs: list[len(list)-1].Ts, NewerTs: list[len(list)-1].Ts}
	raw, _ := json.Marshal(state)
	return s.svcCtx.DataCache.Set(s.ctx, stateKey, raw, 30*24*time.Hour).Err()
}

func (s *GapRepairService) expectedGapJobs(product *models.TMarketProduct, start, end int64) []GapRepairJob {
	category := strings.ToLower(strings.TrimSpace(product.CategoryCode))
	market := strings.ToUpper(strings.TrimSpace(product.Market))
	exchange := strings.TrimSpace(product.Exchange)
	symbol := strings.ToUpper(strings.TrimSpace(product.Symbol))
	var ranges [][2]int64
	var rangeStart int64
	for ts := start; ts <= end; ts += minuteMs {
		if s.ctx.Err() != nil {
			return nil
		}
		trading := s.svcCtx.MarketCalendarResolver == nil || s.svcCtx.MarketCalendarResolver.IsTradingMinute(s.ctx, category, market, exchange, ts)
		if trading && rangeStart == 0 {
			rangeStart = ts
		}
		if !trading && rangeStart > 0 {
			ranges = append(ranges, [2]int64{rangeStart, ts - minuteMs})
			rangeStart = 0
		}
	}
	if rangeStart > 0 {
		ranges = append(ranges, [2]int64{rangeStart, end})
	}
	jobs := make([]GapRepairJob, 0, len(ranges))
	for _, item := range ranges {
		id := gapJobID(category, market, symbol, item[0], item[1])
		jobs = append(jobs, GapRepairJob{ID: id, Category: category, Market: market, Exchange: exchange,
			Symbol: symbol, StartTs: item[0], EndTs: item[1]})
	}
	return jobs
}

func gapJobID(category, market, symbol string, start, end int64) string {
	sum := sha1.Sum([]byte(category + "|" + market + "|" + symbol + "|" + strconv.FormatInt(start, 10) + "|" + strconv.FormatInt(end, 10)))
	return hex.EncodeToString(sum[:])
}

func (s *GapRepairService) enqueue(job GapRepairJob) error {
	if exists, err := s.svcCtx.DataCache.Exists(s.ctx, gapDoneKeyPrefix+job.ID).Result(); err != nil || exists > 0 {
		return err
	}
	if dead, err := s.svcCtx.DataCache.HExists(s.ctx, gapDeadHashKey, job.ID).Result(); err != nil || dead {
		return err
	}
	raw, _ := json.Marshal(job)
	pipe := s.svcCtx.DataCache.TxPipeline()
	pipe.HSetNX(s.ctx, gapJobHashKey, job.ID, raw)
	pipe.ZAddNX(s.ctx, gapQueueKey, redis.Z{Score: float64(time.Now().UnixMilli()), Member: job.ID})
	_, err := pipe.Exec(s.ctx)
	return err
}

func (s *GapRepairService) repairLoop(apiURL, token string) {
	defer s.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			for range s.batchSize {
				job, err := s.claim()
				if err != nil {
					logx.Errorf("claim kline repair job failed: %v", err)
					break
				}
				if job == nil {
					break
				}
				s.repair(apiURL, token, job)
			}
		}
	}
}

func (s *GapRepairService) claim() (*GapRepairJob, error) {
	result, err := claimGapJobScript.Run(s.ctx, s.svcCtx.DataCache, []string{gapQueueKey, gapJobHashKey}, time.Now().UnixMilli()).Slice()
	if err == redis.Nil || len(result) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var raw string
	switch value := result[1].(type) {
	case string:
		raw = value
	case []byte:
		raw = string(value)
	}
	var job GapRepairJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *GapRepairService) repair(apiURL, token string, job *GapRepairJob) {
	// Another overlapping repair or the five-minute reconciliation may already
	// have filled this range while the job was waiting in the retry queue.
	// Remove such stale jobs before making another external request.
	if err := s.verifyGapComplete(job); err == nil {
		s.completeRepair(job, 0, "already complete")
		return
	}

	klineJob := KlineJob{ApiUrl: apiURL, Token: token, Category: job.Category, Market: job.Market,
		Symbol: job.Symbol, KType: reconcileKType}
	result, err := s.worker.syncBackwardAfter(klineJob, "1m", job.EndTs, job.StartTs-minuteMs, 500)
	if err == nil {
		err = s.verifyGapComplete(job)
	}
	if err == nil {
		s.completeRepair(job, result.NewCount, "REST repair")
		return
	}
	job.Attempts++
	if job.Attempts >= gapMaxAttempts {
		s.deadLetter(job, err)
		return
	}
	delay := min(time.Duration(1<<min(job.Attempts, 10))*time.Minute, 12*time.Hour)
	raw, _ := json.Marshal(job)
	pipe := s.svcCtx.DataCache.TxPipeline()
	pipe.HSet(s.ctx, gapJobHashKey, job.ID, raw)
	pipe.ZAdd(s.ctx, gapQueueKey, redis.Z{Score: float64(time.Now().Add(delay).UnixMilli()), Member: job.ID})
	_, _ = pipe.Exec(s.ctx)
	logx.Errorf("repair historical kline gap failed, id=%s category=%s market=%s exchange=%s symbol=%s start=%s end=%s fetched=%d attempts=%d retry=%s err=%v",
		job.ID, job.Category, job.Market, job.Exchange, job.Symbol,
		formatRepairTs(job.StartTs), formatRepairTs(job.EndTs), result.NewCount, job.Attempts, delay, err)
}

func (s *GapRepairService) deadLetter(job *GapRepairJob, repairErr error) {
	letter := gapDeadLetter{Job: job, Error: repairErr.Error(), FailedAt: time.Now().UnixMilli()}
	raw, _ := json.Marshal(letter)
	pipe := s.svcCtx.DataCache.TxPipeline()
	pipe.HDel(s.ctx, gapJobHashKey, job.ID)
	pipe.HSet(s.ctx, gapDeadHashKey, job.ID, raw)
	if _, err := pipe.Exec(s.ctx); err != nil {
		logx.Errorf("dead-letter historical kline gap failed, id=%s err=%v", job.ID, err)
		return
	}
	logx.Errorf("historical kline gap moved to dead-letter after %d attempts, id=%s category=%s market=%s exchange=%s symbol=%s start=%s end=%s err=%v",
		job.Attempts, job.ID, job.Category, job.Market, job.Exchange, job.Symbol,
		formatRepairTs(job.StartTs), formatRepairTs(job.EndTs), repairErr)
}

func (s *GapRepairService) completeRepair(job *GapRepairJob, count int, reason string) {
	pipe := s.svcCtx.DataCache.TxPipeline()
	pipe.HDel(s.ctx, gapJobHashKey, job.ID)
	pipe.HDel(s.ctx, gapDeadHashKey, job.ID)
	pipe.Set(s.ctx, gapDoneKeyPrefix+job.ID, count, 30*24*time.Hour)
	if _, err := pipe.Exec(s.ctx); err != nil {
		logx.Errorf("complete historical kline gap job failed, id=%s err=%v", job.ID, err)
		return
	}
	logx.Infof("completed historical kline gap, category=%s market=%s symbol=%s start=%s end=%s count=%d reason=%s",
		job.Category, job.Market, job.Symbol, formatRepairTs(job.StartTs), formatRepairTs(job.EndTs), count, reason)
}

func (s *GapRepairService) verifyGapComplete(job *GapRepairJob) error {
	model := s.svcCtx.Factory.New(job.Category, "1m")
	if model == nil {
		return fmt.Errorf("invalid 1m model for category=%s", job.Category)
	}
	list, err := model.FindRangeByMarketSymbol(s.ctx, job.Market, job.Symbol, job.StartTs, job.EndTs+minuteMs)
	if err != nil {
		return err
	}
	present := make(map[int64]struct{}, len(list))
	for _, bar := range list {
		if bar != nil {
			present[bar.Ts] = struct{}{}
		}
	}
	missingCount := 0
	var firstMissing, lastMissing int64
	for ts := job.StartTs; ts <= job.EndTs; ts += minuteMs {
		trading := s.svcCtx.MarketCalendarResolver == nil ||
			s.svcCtx.MarketCalendarResolver.IsTradingMinute(s.ctx, job.Category, job.Market, job.Exchange, ts)
		if !trading {
			continue
		}
		if _, ok := present[ts]; !ok {
			if missingCount == 0 {
				firstMissing = ts
			}
			lastMissing = ts
			missingCount++
		}
	}
	if missingCount > 0 {
		return fmt.Errorf("REST completed but MongoDB still has %d missing trading minutes, first=%s last=%s",
			missingCount, formatRepairTs(firstMissing), formatRepairTs(lastMissing))
	}
	return nil
}

func formatRepairTs(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.UnixMilli(ts).UTC().Format(time.RFC3339)
}
