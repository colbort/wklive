package helpers

import (
	"testing"
	"time"

	"wklive/proto/option"
	"wklive/services/option/models"
)

func TestEvaluateTradingCalendarAtContinuousAndBoundaries(t *testing.T) {
	calendar := &models.TOptionTradingCalendar{Timezone: "UTC"}
	sessions := continuousWeeklySessions()

	assertCalendarDecision(t, calendar, sessions, nil, "2026-07-31T23:59:59Z", true, "WEEKLY_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-08-01T00:00:00Z", true, "WEEKLY_SESSION")
}

func TestEvaluateTradingCalendarAtDSTAndCrossMidnight(t *testing.T) {
	calendar := &models.TOptionTradingCalendar{Timezone: "America/New_York"}
	sessions := []*models.TOptionTradingCalendarSession{
		{Weekday: int64(time.Monday), OpenSecond: 9*3600 + 30*60, CloseSecond: 16 * 3600},
	}

	assertCalendarDecision(t, calendar, sessions, nil, "2026-03-02T14:29:59Z", false, "OUTSIDE_TRADING_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-03-02T14:30:00Z", true, "WEEKLY_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-03-09T13:30:00Z", true, "WEEKLY_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-03-09T20:00:00Z", false, "OUTSIDE_TRADING_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-10-26T13:30:00Z", true, "WEEKLY_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-11-02T14:29:59Z", false, "OUTSIDE_TRADING_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-11-02T14:30:00Z", true, "WEEKLY_SESSION")

	calendar.Timezone = "UTC"
	sessions = []*models.TOptionTradingCalendarSession{
		{Weekday: int64(time.Monday), OpenSecond: 22 * 3600, CloseSecond: 26 * 3600},
	}
	assertCalendarDecision(t, calendar, sessions, nil, "2026-08-03T22:00:00Z", true, "WEEKLY_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-08-04T01:59:59Z", true, "WEEKLY_SESSION")
	assertCalendarDecision(t, calendar, sessions, nil, "2026-08-04T02:00:00Z", false, "OUTSIDE_TRADING_SESSION")
}

func TestEvaluateTradingCalendarExceptionPrecedence(t *testing.T) {
	calendar := &models.TOptionTradingCalendar{Timezone: "UTC"}
	at := mustParseCalendarTime(t, "2026-08-03T12:00:00Z")
	openException := &models.TOptionTradingCalendarException{
		ExceptionType: int64(option.TradingCalendarExceptionType_TRADING_CALENDAR_EXCEPTION_TYPE_OPEN),
		StartTime:     at.Add(-time.Hour).Unix(),
		EndTime:       at.Add(time.Hour).Unix(),
	}
	closedException := &models.TOptionTradingCalendarException{
		ExceptionType: int64(option.TradingCalendarExceptionType_TRADING_CALENDAR_EXCEPTION_TYPE_CLOSED),
		StartTime:     at.Add(-time.Hour).Unix(),
		EndTime:       at.Add(time.Hour).Unix(),
	}

	decision, err := EvaluateTradingCalendarAt(calendar, nil, []*models.TOptionTradingCalendarException{openException}, at)
	if err != nil || !decision.Open || decision.Reason != "CALENDAR_OPEN_EXCEPTION" {
		t.Fatalf("OPEN exception did not override closed weekly schedule: decision=%+v err=%v", decision, err)
	}
	decision, err = EvaluateTradingCalendarAt(
		calendar, continuousWeeklySessions(),
		[]*models.TOptionTradingCalendarException{openException, closedException}, at,
	)
	if err != nil || decision.Open || decision.Reason != "CALENDAR_CLOSED_EXCEPTION" {
		t.Fatalf("CLOSED exception did not win: decision=%+v err=%v", decision, err)
	}
}

func TestValidateTradingCalendarDefinition(t *testing.T) {
	valid := []*models.TOptionTradingCalendarSession{
		{Weekday: int64(time.Monday), OpenSecond: 22 * 3600, CloseSecond: 26 * 3600},
	}
	if err := ValidateTradingCalendarDefinition("Asia/Hong_Kong", 100, 0, valid, nil); err != nil {
		t.Fatalf("valid calendar rejected: %v", err)
	}
	if err := ValidateTradingCalendarDefinition("Mars/Olympus", 100, 0, valid, nil); err == nil {
		t.Fatal("invalid timezone accepted")
	}
	overlap := append(valid, &models.TOptionTradingCalendarSession{
		Weekday: int64(time.Tuesday), OpenSecond: 3600, CloseSecond: 3 * 3600,
	})
	if err := ValidateTradingCalendarDefinition("UTC", 100, 0, overlap, nil); err == nil {
		t.Fatal("cross-midnight weekly overlap accepted")
	}
	exceptions := []*models.TOptionTradingCalendarException{
		{ExceptionType: 1, StartTime: 110, EndTime: 130, Reason: "holiday"},
		{ExceptionType: 1, StartTime: 120, EndTime: 140, Reason: "maintenance"},
	}
	if err := ValidateTradingCalendarDefinition("UTC", 100, 0, valid, exceptions); err == nil {
		t.Fatal("same-type exception overlap accepted")
	}
}

func continuousWeeklySessions() []*models.TOptionTradingCalendarSession {
	sessions := make([]*models.TOptionTradingCalendarSession, 0, 7)
	for weekday := int64(0); weekday < 7; weekday++ {
		sessions = append(sessions, &models.TOptionTradingCalendarSession{
			Weekday: weekday, OpenSecond: 0, CloseSecond: 86400,
		})
	}
	return sessions
}

func assertCalendarDecision(
	t *testing.T,
	calendar *models.TOptionTradingCalendar,
	sessions []*models.TOptionTradingCalendarSession,
	exceptions []*models.TOptionTradingCalendarException,
	at string,
	wantOpen bool,
	wantReason string,
) {
	t.Helper()
	decision, err := EvaluateTradingCalendarAt(calendar, sessions, exceptions, mustParseCalendarTime(t, at))
	if err != nil {
		t.Fatalf("evaluate %s: %v", at, err)
	}
	if decision.Open != wantOpen || decision.Reason != wantReason {
		t.Fatalf("evaluate %s: got %+v, want open=%v reason=%s", at, decision, wantOpen, wantReason)
	}
}

func mustParseCalendarTime(t *testing.T, value string) time.Time {
	t.Helper()
	result, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse test time: %v", err)
	}
	return result
}
