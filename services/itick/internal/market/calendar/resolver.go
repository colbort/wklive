package calendar

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"wklive/services/itick/models"
)

type Definition struct {
	ID               int64
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
	mu       sync.RWMutex
	cache    map[string]cacheItem
	holidays map[string]holidayItem
	ttl      time.Duration
}

func NewResolver(model Model, ttl time.Duration) *Resolver {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Resolver{model: model, cache: make(map[string]cacheItem), holidays: make(map[string]holidayItem), ttl: ttl}
}

// IsTradingMinute reports whether a missing 1m bucket is expected to exist.
// Crypto falls back to 24x7. Other categories require calendar sessions; this
// conservative fallback prevents overnight/weekend false-positive repairs.
func (r *Resolver) IsTradingMinute(ctx context.Context, category, market, exchange string, ts int64) bool {
	d := r.Resolve(ctx, category, market, exchange)
	if d.ID == 0 {
		return strings.EqualFold(category, "crypto")
	}
	local := time.UnixMilli(ts).In(d.Location)
	date := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, d.Location)
	holiday := r.holiday(ctx, d.ID, date)
	if holiday != nil {
		if strings.EqualFold(holiday.DayType, "closed") {
			return false
		}
		if holiday.OpenTime != "" && holiday.CloseTime != "" {
			return withinClock(local, holiday.OpenTime, holiday.CloseTime, false)
		}
	}
	if len(d.Sessions) == 0 {
		return strings.EqualFold(category, "crypto")
	}
	for _, session := range d.Sessions {
		if session != nil && withinClock(local, session.StartTime, session.EndTime, session.CrossDay != 0) {
			day := local.Weekday()
			if session.CrossDay != 0 {
				_, end, ok := clockMinutes(session.StartTime, session.EndTime)
				if ok && local.Hour()*60+local.Minute() < end {
					day = local.AddDate(0, 0, -1).Weekday()
				}
			}
			return day != time.Saturday && day != time.Sunday
		}
	}
	return false
}

func (r *Resolver) holiday(ctx context.Context, calendarID int64, date time.Time) *models.TItickMarketHoliday {
	key := fmt.Sprintf("%d:%s", calendarID, date.Format("2006-01-02"))
	r.mu.RLock()
	item, ok := r.holidays[key]
	r.mu.RUnlock()
	if ok && time.Now().Before(item.expires) {
		return item.value
	}
	var value *models.TItickMarketHoliday
	if r.model != nil {
		value, _ = r.model.FindHoliday(ctx, calendarID, date)
	}
	r.mu.Lock()
	r.holidays[key] = holidayItem{value: value, expires: time.Now().Add(r.ttl)}
	r.mu.Unlock()
	return value
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
	parse := func(value string) (int, bool) {
		parts := strings.Split(strings.TrimSpace(value), ":")
		if len(parts) < 2 {
			return 0, false
		}
		hour, err1 := strconv.Atoi(parts[0])
		minute, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
			return 0, false
		}
		return hour*60 + minute, true
	}
	start, ok1 := parse(startValue)
	end, ok2 := parse(endValue)
	return start, end, ok1 && ok2
}

func (r *Resolver) Resolve(ctx context.Context, category, market, exchange string) *Definition {
	key := strings.ToLower(category) + ":" + strings.ToUpper(market) + ":" + strings.ToUpper(exchange)
	r.mu.RLock()
	item, ok := r.cache[key]
	r.mu.RUnlock()
	if ok && time.Now().Before(item.expires) {
		return item.value
	}
	def := &Definition{Location: time.UTC, WeekStart: time.Monday}
	if r.model != nil {
		if row, err := r.model.Resolve(ctx, category, market, exchange); err == nil && row != nil {
			def.ID, def.TradingDayOffset = row.Id, int(row.TradingDayOffset)
			if loc, loadErr := time.LoadLocation(row.Timezone); loadErr == nil {
				def.Location = loc
			}
			if row.WeekStart >= 1 && row.WeekStart <= 7 {
				def.WeekStart = time.Weekday(row.WeekStart % 7)
			}
			def.Sessions, _ = r.model.FindSessions(ctx, row.Id)
		}
	}
	r.mu.Lock()
	r.cache[key] = cacheItem{def, time.Now().Add(r.ttl)}
	r.mu.Unlock()
	return def
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
