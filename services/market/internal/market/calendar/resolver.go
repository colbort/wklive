package calendar

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"wklive/services/market/models"
)

type Definition struct {
	ID               int64
	CategoryCode     string
	Market           string
	Exchange         string
	Timezone         string
	ProductSpecific  bool
	Remark           string
	Location         *time.Location
	TradingDayOffset int
	WeekStart        time.Weekday
	Sessions         []*models.TItickMarketSession
}

type Model interface {
	Resolve(context.Context, string, string, string) (*models.TItickMarketCalendar, error)
	FindSessions(context.Context, int64) ([]*models.TItickMarketSession, error)
	FindHoliday(context.Context, int64, time.Time) (*models.TItickMarketHoliday, error)
}

type ProductCalendarModel interface {
	ResolveCalendar(context.Context, string, string, string) (*models.TItickMarketCalendar, error)
}

type cacheItem struct {
	value   *Definition
	expires time.Time
}

type holidayItem struct {
	value   *models.TItickMarketHoliday
	expires time.Time
}

type Resolver struct {
	model    Model
	products ProductCalendarModel
	mu       sync.RWMutex
	cache    map[string]cacheItem
	holidays map[string]holidayItem
	ttl      time.Duration
}

func NewResolver(model Model, products ProductCalendarModel, ttl time.Duration) *Resolver {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Resolver{model: model, products: products, cache: make(map[string]cacheItem), holidays: make(map[string]holidayItem), ttl: ttl}
}

func (r *Resolver) Invalidate() {
	r.mu.Lock()
	clear(r.cache)
	clear(r.holidays)
	r.mu.Unlock()
}

// IsTradingMinute reports whether a missing 1m bucket is expected to exist.
// Crypto falls back to 24x7. Other categories require calendar sessions; this
// conservative fallback prevents overnight/weekend false-positive repairs.
func (r *Resolver) IsTradingMinute(ctx context.Context, category, market, exchange string, ts int64) bool {
	return r.isTradingMinute(ctx, r.Resolve(ctx, category, market, exchange), category, ts)
}

func (r *Resolver) IsProductTradingMinute(ctx context.Context, productID int64, category, market, symbol, exchange string, ts int64) bool {
	return r.isTradingMinute(ctx, r.ResolveProduct(ctx, productID, category, market, symbol, exchange), category, ts)
}

func (r *Resolver) IsSymbolTradingMinute(ctx context.Context, category, market, symbol, exchange string, ts int64) bool {
	return r.isTradingMinute(ctx, r.ResolveSymbol(ctx, category, market, symbol, exchange), category, ts)
}

func (r *Resolver) IsResolvedTradingMinute(ctx context.Context, definition *Definition, category string, ts int64) bool {
	open, _ := r.EvaluateResolvedTradingMinute(ctx, definition, category, ts)
	return open
}

func (r *Resolver) EvaluateResolvedTradingMinute(ctx context.Context, definition *Definition, category string, ts int64) (bool, error) {
	open, _, err := r.EvaluateResolvedTradingSession(ctx, definition, category, ts)
	return open, err
}

// EvaluateResolvedTradingSession reports whether the market is open and which
// configured session matched the requested timestamp.
func (r *Resolver) EvaluateResolvedTradingSession(ctx context.Context, definition *Definition, category string, ts int64) (bool, string, error) {
	if definition == nil {
		return false, "", errors.New("market calendar definition is unavailable")
	}
	return r.evaluateTradingSession(ctx, definition, category, ts)
}

func (r *Resolver) isTradingMinute(ctx context.Context, d *Definition, category string, ts int64) bool {
	open, _ := r.evaluateTradingMinute(ctx, d, category, ts)
	return open
}

func (r *Resolver) evaluateTradingMinute(ctx context.Context, d *Definition, category string, ts int64) (bool, error) {
	open, _, err := r.evaluateTradingSession(ctx, d, category, ts)
	return open, err
}

func (r *Resolver) evaluateTradingSession(ctx context.Context, d *Definition, category string, ts int64) (bool, string, error) {
	if d.ID == 0 {
		if strings.EqualFold(category, "crypto") {
			return true, "24x7", nil
		}
		return false, "", nil
	}
	local := time.UnixMilli(ts).In(d.Location)
	date := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, d.Location)
	holiday, err := r.holiday(ctx, d.ID, date)
	if err != nil {
		return false, "", err
	}
	if holiday != nil {
		if strings.EqualFold(holiday.DayType, "closed") {
			return false, "", nil
		}
		if holiday.OpenTime != "" && holiday.CloseTime != "" {
			if withinClock(local, holiday.OpenTime, holiday.CloseTime, false) {
				return true, "regular", nil
			}
			return false, "", nil
		}
	}
	if len(d.Sessions) == 0 {
		if strings.EqualFold(category, "crypto") {
			return true, "24x7", nil
		}
		return false, "", nil
	}
	for _, session := range d.Sessions {
		if sessionActive(local, session) {
			return true, strings.ToLower(strings.TrimSpace(session.SessionType)), nil
		}
	}
	return false, "", nil
}

func sessionActive(local time.Time, session *models.TItickMarketSession) bool {
	if session == nil || !withinClock(local, session.StartTime, session.EndTime, session.CrossDay != 0) {
		return false
	}
	day := local.Weekday()
	if session.CrossDay != 0 {
		_, end, ok := clockMinutes(session.StartTime, session.EndTime)
		if ok && local.Hour()*60+local.Minute() < end {
			day = local.AddDate(0, 0, -1).Weekday()
		}
	}
	mask := session.WeekdayMask
	if mask == 0 {
		mask = 62 // Backward-compatible Monday-Friday default.
	}
	return mask&(1<<uint(day)) != 0
}

func (r *Resolver) holiday(ctx context.Context, calendarID int64, date time.Time) (*models.TItickMarketHoliday, error) {
	key := fmt.Sprintf("%d:%s", calendarID, date.Format("2006-01-02"))
	r.mu.RLock()
	item, ok := r.holidays[key]
	r.mu.RUnlock()
	if ok && time.Now().Before(item.expires) {
		return item.value, nil
	}
	var value *models.TItickMarketHoliday
	if r.model != nil {
		var err error
		value, err = r.model.FindHoliday(ctx, calendarID, date)
		if err != nil {
			return nil, err
		}
	}
	r.mu.Lock()
	r.holidays[key] = holidayItem{value: value, expires: time.Now().Add(r.ttl)}
	r.mu.Unlock()
	return value, nil
}

func withinClock(t time.Time, startValue, endValue string, crossDay bool) bool {
	start, end, ok := clockMinutes(startValue, endValue)
	if !ok {
		return false
	}
	minute := t.Hour()*60 + t.Minute()
	if crossDay || end <= start {
		return minute >= start || minute < end
	}
	return minute >= start && minute < end
}

func clockMinutes(startValue, endValue string) (int, int, bool) {
	parse := func(value string, allowEndOfDay bool) (int, bool) {
		parts := strings.Split(strings.TrimSpace(value), ":")
		if len(parts) < 2 {
			return 0, false
		}
		hour, err1 := strconv.Atoi(parts[0])
		minute, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || hour < 0 || minute < 0 || minute > 59 || hour > 24 || (hour == 24 && (!allowEndOfDay || minute != 0)) {
			return 0, false
		}
		return hour*60 + minute, true
	}
	start, ok1 := parse(startValue, false)
	end, ok2 := parse(endValue, true)
	return start, end, ok1 && ok2
}

func (r *Resolver) Resolve(ctx context.Context, category, market, exchange string) *Definition {
	key := strings.ToLower(category) + ":" + strings.ToUpper(market) + ":" + strings.ToUpper(exchange)
	definition, _ := r.resolve(ctx, key, nil, false, category, market, exchange)
	return definition
}

func (r *Resolver) ResolveProduct(ctx context.Context, productID int64, category, market, symbol, exchange string) *Definition {
	key := fmt.Sprintf("product:%d", productID)
	definition, _ := r.resolveMapped(ctx, key, category, market, symbol, exchange)
	return definition
}

func (r *Resolver) ResolveSymbol(ctx context.Context, category, market, symbol, exchange string) *Definition {
	definition, _ := r.ResolveSymbolStrict(ctx, category, market, symbol, exchange)
	return definition
}

func (r *Resolver) ResolveSymbolStrict(ctx context.Context, category, market, symbol, exchange string) (*Definition, error) {
	key := "symbol:" + strings.ToLower(strings.TrimSpace(category)) + ":" +
		strings.ToUpper(strings.TrimSpace(market)) + ":" + strings.ToUpper(strings.TrimSpace(symbol)) + ":" +
		strings.ToUpper(strings.TrimSpace(exchange))
	return r.resolveMapped(ctx, key, category, market, symbol, exchange)
}

func (r *Resolver) resolveMapped(ctx context.Context, key, category, market, symbol, exchange string) (*Definition, error) {
	r.mu.RLock()
	item, ok := r.cache[key]
	r.mu.RUnlock()
	if ok && time.Now().Before(item.expires) {
		return item.value, nil
	}
	var row *models.TItickMarketCalendar
	productSpecific := false
	if r.products != nil && strings.TrimSpace(symbol) != "" {
		var err error
		row, err = r.products.ResolveCalendar(ctx, category, market, symbol)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		productSpecific = row != nil
	}
	return r.resolve(ctx, key, row, productSpecific, category, market, exchange)
}

func (r *Resolver) resolve(ctx context.Context, key string, selected *models.TItickMarketCalendar, productSpecific bool, category, market, exchange string) (*Definition, error) {
	r.mu.RLock()
	item, ok := r.cache[key]
	r.mu.RUnlock()
	if ok && time.Now().Before(item.expires) {
		return item.value, nil
	}
	def := &Definition{Location: time.UTC, Timezone: "UTC", WeekStart: time.Monday, ProductSpecific: productSpecific}
	if r.model != nil {
		row := selected
		if row == nil {
			var err error
			row, err = r.model.Resolve(ctx, category, market, exchange)
			if err != nil && !errors.Is(err, models.ErrNotFound) {
				return nil, err
			}
		}
		if row != nil {
			def.ID, def.CategoryCode, def.Market, def.Exchange = row.Id, row.CategoryCode, row.Market, row.Exchange
			def.Timezone, def.Remark, def.TradingDayOffset = row.Timezone, row.Remark, int(row.TradingDayOffset)
			if loc, loadErr := time.LoadLocation(row.Timezone); loadErr == nil {
				def.Location = loc
			}
			if row.WeekStart >= 1 && row.WeekStart <= 7 {
				def.WeekStart = time.Weekday(row.WeekStart % 7)
			}
			var err error
			def.Sessions, err = r.model.FindSessions(ctx, row.Id)
			if err != nil {
				return nil, err
			}
		}
	}
	r.mu.Lock()
	r.cache[key] = cacheItem{def, time.Now().Add(r.ttl)}
	r.mu.Unlock()
	return def, nil
}

func (r *Resolver) Bucket(ctx context.Context, category, market, exchange string, ts int64, interval string) (int64, int64) {
	if interval == "5m" || interval == "15m" || interval == "30m" || interval == "1h" {
		minutes := map[string]int64{"5m": 5, "15m": 15, "30m": 30, "1h": 60}[interval]
		width := minutes * 60_000
		start := ts / width * width
		return start, start + width
	}
	d := r.Resolve(ctx, category, market, exchange)
	local := time.UnixMilli(ts).In(d.Location).AddDate(0, 0, d.TradingDayOffset)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, d.Location)
	var start, end time.Time
	switch interval {
	case "1w":
		offset := (7 + int(day.Weekday()) - int(d.WeekStart)) % 7
		start, end = day.AddDate(0, 0, -offset), day.AddDate(0, 0, -offset+7)
	case "1mo":
		start = time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, d.Location)
		end = start.AddDate(0, 1, 0)
	case "1y":
		start = time.Date(local.Year(), 1, 1, 0, 0, 0, 0, d.Location)
		end = start.AddDate(1, 0, 0)
	default:
		start, end = day, day.AddDate(0, 0, 1)
	}
	start, end = start.AddDate(0, 0, -d.TradingDayOffset), end.AddDate(0, 0, -d.TradingDayOffset)
	return start.UTC().UnixMilli(), end.UTC().UnixMilli()
}
