package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"wklive/services/market/internal/pkg/itickrest"
	"wklive/services/market/internal/pkg/utils"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	holidaySyncLockKey = "market:market_calendar:holiday_sync"
	holidaySyncLockTTL = 30 * time.Minute
)

type HolidaySyncService struct {
	ctx        context.Context
	cancel     context.CancelFunc
	apiURL     string
	calendar   models.TItickMarketCalendarModel
	holiday    models.TItickMarketHolidayModel
	resolver   *Resolver
	lock       *utils.RedisLock
	restClient *itickrest.Client
	interval   time.Duration
	wg         sync.WaitGroup
}

type holidayResponse struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data []holidayAPIItem `json:"data"`
}

type holidayAPIItem struct {
	Code         string `json:"c"`
	Region       string `json:"r"`
	Date         string `json:"d"`
	TradingHours string `json:"t"`
	Timezone     string `json:"z"`
	Name         string `json:"v"`
}

func NewHolidaySyncService(
	parent context.Context,
	apiURL string,
	restClient *itickrest.Client,
	calendar models.TItickMarketCalendarModel,
	holiday models.TItickMarketHolidayModel,
	resolver *Resolver,
	lock *utils.RedisLock,
	interval time.Duration,
) *HolidaySyncService {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	ctx, cancel := context.WithCancel(parent)
	return &HolidaySyncService{
		ctx:        ctx,
		cancel:     cancel,
		apiURL:     strings.TrimSpace(apiURL),
		calendar:   calendar,
		holiday:    holiday,
		resolver:   resolver,
		lock:       lock,
		interval:   interval,
		restClient: restClient,
	}
}

func (s *HolidaySyncService) Start() {
	s.wg.Add(1)
	go s.loop()
}

func (s *HolidaySyncService) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *HolidaySyncService) loop() {
	defer s.wg.Done()
	s.run()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.run()
		}
	}
}

func (s *HolidaySyncService) run() {
	if s.apiURL == "" || s.restClient == nil || s.calendar == nil || s.holiday == nil {
		logx.Error("skip iTick holiday sync: missing REST client, api URL or model")
		return
	}
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	if s.lock != nil {
		if err := s.lock.Acquire(s.ctx, holidaySyncLockKey, lockValue, holidaySyncLockTTL); err != nil {
			if !errors.Is(err, utils.ErrLockNotAcquired) {
				logx.Errorf("acquire iTick holiday sync lock failed: %v", err)
			}
			return
		}
		defer func() {
			if err := s.lock.Release(context.Background(), holidaySyncLockKey, lockValue); err != nil {
				logx.Errorf("release iTick holiday sync lock failed: %v", err)
			}
		}()
	}
	count, err := s.SyncOnce(s.ctx)
	if count > 0 && s.resolver != nil {
		s.resolver.Invalidate()
	}
	if err != nil {
		logx.Errorf("sync iTick market holidays partial failure, rows=%d err=%v", count, err)
		return
	}
	logx.Infof("synced iTick market holidays, rows=%d", count)
}

func (s *HolidaySyncService) SyncOnce(ctx context.Context) (int, error) {
	byCode := utils.StockHolidayMarketsByCode()
	codes := make([]string, 0, len(byCode))
	for code := range byCode {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	total := 0
	var failures []error
	for _, code := range codes {
		items, err := s.fetch(ctx, code)
		if err != nil {
			failures = append(failures, fmt.Errorf("code=%s: %w", code, err))
			continue
		}
		if len(items) == 0 {
			continue
		}
		timezone := normalizeHolidayTimezone(items[0].Timezone)
		if timezone == "" {
			timezone = "UTC"
		}
		if _, err := time.LoadLocation(timezone); err != nil {
			failures = append(failures, fmt.Errorf("code=%s invalid timezone %q: %w", code, timezone, err))
			continue
		}
		for _, market := range byCode[code] {
			calendar, err := s.calendar.Ensure(ctx, "stock", market, "", timezone, time.Now().UnixMilli())
			if err != nil {
				failures = append(failures, fmt.Errorf("code=%s market=%s ensure calendar: %w", code, market, err))
				continue
			}
			for _, item := range items {
				tradeDate, err := time.ParseInLocation("2006-01-02", item.Date, time.UTC)
				if err != nil {
					failures = append(failures, fmt.Errorf("code=%s invalid date %q: %w", code, item.Date, err))
					continue
				}
				if err := s.holiday.UpsertByCalendarDate(ctx, &models.TItickMarketHoliday{
					CalendarId: calendar.Id, TradeDate: tradeDate, DayType: "closed", Remark: strings.TrimSpace(item.Name),
				}); err != nil {
					failures = append(failures, fmt.Errorf("code=%s market=%s date=%s: %w", code, market, item.Date, err))
					continue
				}
				total++
			}
		}
	}
	return total, errors.Join(failures...)
}

func normalizeHolidayTimezone(value string) string {
	value = strings.TrimSpace(value)
	aliases := map[string]string{
		"Asia/Mumbai": "Asia/Kolkata",
	}
	if canonical, ok := aliases[value]; ok {
		return canonical
	}
	return value
}

func (s *HolidaySyncService) fetch(ctx context.Context, code string) ([]holidayAPIItem, error) {
	base, err := url.Parse(s.apiURL)
	if err != nil {
		return nil, err
	}
	base.Path = path.Join(strings.TrimRight(base.Path, "/"), "symbol/v2/holidays")
	query := base.Query()
	query.Set("code", code)
	base.RawQuery = query.Encode()
	resp, err := s.restClient.Get(ctx, base.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out holidayResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("REST rejected: code=%d msg=%s", out.Code, out.Msg)
	}
	return out.Data, nil
}
