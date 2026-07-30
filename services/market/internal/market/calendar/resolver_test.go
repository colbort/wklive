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
	r := NewResolver(stub, time.Minute)
	ts := time.Date(2026, 7, 14, 14, 0, 0, 0, time.UTC).UnixMilli()
	if r.IsTradingMinute(context.Background(), "stock", "US", "", ts) {
		t.Fatal("closed holiday must not produce a repair gap")
	}
}

func TestBucketUsesMarketTimezone(t *testing.T) {
	r := NewResolver(calendarModelStub{row: &models.TItickMarketCalendar{Id: 1, Timezone: "America/New_York", WeekStart: 1}}, time.Minute)
	ts := time.Date(2026, 7, 14, 16, 0, 0, 0, time.UTC).UnixMilli()
	start, end := r.Bucket(context.Background(), "stock", "US", "", ts, "1d")
	want := time.Date(2026, 7, 14, 4, 0, 0, 0, time.UTC).UnixMilli()
	if start != want || end-start != int64(24*time.Hour/time.Millisecond) {
		t.Fatalf("unexpected day bucket %d-%d", start, end)
	}
}

func TestBucketFallsBackToUTC(t *testing.T) {
	r := NewResolver(nil, time.Minute)
	ts := time.Date(2026, 7, 14, 16, 0, 0, 0, time.UTC).UnixMilli()
	start, _ := r.Bucket(context.Background(), "crypto", "BA", "", ts, "1d")
	if time.UnixMilli(start).UTC().Hour() != 0 {
		t.Fatalf("expected UTC midnight, got %v", time.UnixMilli(start))
	}
}

func TestIsTradingMinuteUsesSession(t *testing.T) {
	stub := calendarModelStub{row: &models.TItickMarketCalendar{Id: 1, Timezone: "America/New_York", WeekStart: 1}}
	stub.sessions = []*models.TItickMarketSession{{StartTime: "09:30", EndTime: "16:00"}}
	r := NewResolver(stub, time.Minute)
	if !r.IsTradingMinute(context.Background(), "stock", "US", "", time.Date(2026, 7, 14, 14, 0, 0, 0, time.UTC).UnixMilli()) {
		t.Fatal("expected regular session minute")
	}
	if r.IsTradingMinute(context.Background(), "stock", "US", "", time.Date(2026, 7, 14, 22, 0, 0, 0, time.UTC).UnixMilli()) {
		t.Fatal("expected after-hours minute to be closed")
	}
}
