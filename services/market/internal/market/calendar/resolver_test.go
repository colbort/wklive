package calendar

import (
	"context"
	"testing"
	"time"

	"wklive/services/market/models"
)

type calendarModelStub struct {
	row      *models.TItickMarketCalendar
	sessions []*models.TItickMarketSession
	holiday  *models.TItickMarketHoliday
}

func (s calendarModelStub) Resolve(context.Context, string, string, string) (*models.TItickMarketCalendar, error) {
	return s.row, nil
}
func (s calendarModelStub) FindSessions(context.Context, int64) ([]*models.TItickMarketSession, error) {
	return s.sessions, nil
}
func (s calendarModelStub) FindHoliday(context.Context, int64, time.Time) (*models.TItickMarketHoliday, error) {
	return s.holiday, nil
}

func TestIsTradingMinuteHonorsClosedHoliday(t *testing.T) {
	stub := calendarModelStub{
		row:      &models.TItickMarketCalendar{Id: 1, Timezone: "America/New_York", WeekStart: 1},
		sessions: []*models.TItickMarketSession{{StartTime: "09:30", EndTime: "16:00"}},
		holiday:  &models.TItickMarketHoliday{DayType: "closed"},
	}
	r := NewResolver(stub, nil, time.Minute)
	ts := time.Date(2026, 7, 14, 14, 0, 0, 0, time.UTC).UnixMilli()
	if r.IsTradingMinute(context.Background(), "stock", "US", "", ts) {
		t.Fatal("closed holiday must not produce a repair gap")
	}
}

func TestBucketUsesMarketTimezone(t *testing.T) {
	r := NewResolver(calendarModelStub{row: &models.TItickMarketCalendar{Id: 1, Timezone: "America/New_York", WeekStart: 1}}, nil, time.Minute)
	ts := time.Date(2026, 7, 14, 16, 0, 0, 0, time.UTC).UnixMilli()
	start, end := r.Bucket(context.Background(), "stock", "US", "", ts, "1d")
	want := time.Date(2026, 7, 14, 4, 0, 0, 0, time.UTC).UnixMilli()
	if start != want || end-start != int64(24*time.Hour/time.Millisecond) {
		t.Fatalf("unexpected day bucket %d-%d", start, end)
	}
}

func TestBucketFallsBackToUTC(t *testing.T) {
	r := NewResolver(nil, nil, time.Minute)
	ts := time.Date(2026, 7, 14, 16, 0, 0, 0, time.UTC).UnixMilli()
	start, _ := r.Bucket(context.Background(), "crypto", "BA", "", ts, "1d")
	if time.UnixMilli(start).UTC().Hour() != 0 {
		t.Fatalf("expected UTC midnight, got %v", time.UnixMilli(start))
	}
}

func TestIsTradingMinuteUsesSession(t *testing.T) {
	stub := calendarModelStub{row: &models.TItickMarketCalendar{Id: 1, Timezone: "America/New_York", WeekStart: 1}}
	stub.sessions = []*models.TItickMarketSession{{SessionType: "regular", StartTime: "09:30", EndTime: "16:00"}}
	r := NewResolver(stub, nil, time.Minute)
	if !r.IsTradingMinute(context.Background(), "stock", "US", "", time.Date(2026, 7, 14, 14, 0, 0, 0, time.UTC).UnixMilli()) {
		t.Fatal("expected regular session minute")
	}
	if r.IsTradingMinute(context.Background(), "stock", "US", "", time.Date(2026, 7, 14, 22, 0, 0, 0, time.UTC).UnixMilli()) {
		t.Fatal("expected after-hours minute to be closed")
	}
}

func TestEvaluateResolvedTradingSessionReturnsMatchedType(t *testing.T) {
	stub := calendarModelStub{
		row: &models.TItickMarketCalendar{Id: 1, Timezone: "America/New_York", WeekStart: 1},
		sessions: []*models.TItickMarketSession{
			{SessionType: "pre", StartTime: "04:00", EndTime: "09:30", WeekdayMask: 62},
			{SessionType: "regular", StartTime: "09:30", EndTime: "16:00", WeekdayMask: 62},
		},
	}
	r := NewResolver(stub, nil, time.Minute)
	definition := r.Resolve(context.Background(), "stock", "US", "")
	open, sessionType, err := r.EvaluateResolvedTradingSession(
		context.Background(), definition, "stock",
		time.Date(2026, 7, 14, 9, 0, 0, 0, time.FixedZone("EDT", -4*60*60)).UnixMilli(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !open || sessionType != "pre" {
		t.Fatalf("unexpected trading session: open=%v session=%q", open, sessionType)
	}
}

func TestIsTradingMinuteHonorsSundayAndCrossDay(t *testing.T) {
	stub := calendarModelStub{
		row: &models.TItickMarketCalendar{Id: 1, Timezone: "America/New_York", WeekStart: 1},
		sessions: []*models.TItickMarketSession{{
			StartTime: "17:05", EndTime: "16:59", CrossDay: 1, WeekdayMask: 31,
		}},
	}
	r := NewResolver(stub, nil, time.Minute)
	// Sunday 18:00 in New York belongs to the Sunday-started session.
	if !r.IsTradingMinute(context.Background(), "forex", "GB", "", time.Date(2026, 7, 12, 22, 0, 0, 0, time.UTC).UnixMilli()) {
		t.Fatal("expected Sunday forex session to be open")
	}
	// Friday after the daily close must be closed until Sunday evening.
	if r.IsTradingMinute(context.Background(), "forex", "GB", "", time.Date(2026, 7, 10, 22, 0, 0, 0, time.UTC).UnixMilli()) {
		t.Fatal("expected Friday maintenance/weekend window to be closed")
	}
}

type productCalendarStub struct {
	row *models.TItickMarketCalendar
}

func (s productCalendarStub) ResolveCalendar(context.Context, string, string, string) (*models.TItickMarketCalendar, error) {
	return s.row, nil
}

func TestProductCalendarOverridesMarketFallback(t *testing.T) {
	market := calendarModelStub{
		row:      &models.TItickMarketCalendar{Id: 1, Timezone: "UTC", WeekStart: 1},
		sessions: []*models.TItickMarketSession{{StartTime: "09:00", EndTime: "10:00", WeekdayMask: 62}},
	}
	product := productCalendarStub{row: &models.TItickMarketCalendar{
		Id: 2, CategoryCode: "future", Market: "US", Exchange: "CME",
		Timezone: "UTC", WeekStart: 1, Remark: "ES calendar",
	}}
	resolver := NewResolver(productAwareCalendarModel{calendarModelStub: market}, product, time.Minute)
	definition := resolver.ResolveProduct(context.Background(), 99, "future", "US", "ES", "CME")
	if definition.ID != 2 || !definition.ProductSpecific || definition.Market != "US" || definition.Exchange != "CME" {
		t.Fatalf("unexpected product calendar definition: %+v", definition)
	}
	if len(definition.Sessions) != 1 || definition.Sessions[0].StartTime != "17:00" {
		t.Fatalf("unexpected product calendar sessions: %+v", definition.Sessions)
	}
	if !resolver.IsProductTradingMinute(context.Background(), 99, "future", "US", "ES", "CME", time.Date(2026, 7, 13, 18, 0, 0, 0, time.UTC).UnixMilli()) {
		t.Fatal("expected product-specific afternoon session")
	}
}

type productAwareCalendarModel struct {
	calendarModelStub
}

func (s productAwareCalendarModel) FindSessions(_ context.Context, calendarID int64) ([]*models.TItickMarketSession, error) {
	if calendarID == 2 {
		return []*models.TItickMarketSession{{StartTime: "17:00", EndTime: "16:00", CrossDay: 1, WeekdayMask: 31}}, nil
	}
	return s.calendarModelStub.FindSessions(context.Background(), calendarID)
}
