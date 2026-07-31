package helpers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

const DefaultTradingCalendarCode = "CONTINUOUS_24_7"

var tradingCalendarCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{0,63}$`)

type TradingCalendarDecision struct {
	Open       bool
	Reason     string
	CalendarId int64
	Version    int64
}

func NormalizeTradingCalendarCode(value string) (string, bool) {
	value = strings.ToUpper(strings.TrimSpace(value))
	return value, tradingCalendarCodePattern.MatchString(value)
}

func IsContractTradingOpen(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	now int64,
) (*TradingCalendarDecision, error) {
	return IsContractTradingOpenWithModels(
		ctx,
		svcCtx.OptionTradingHaltModel,
		svcCtx.OptionTradingCalendarModel,
		svcCtx.OptionTradingCalendarSessionModel,
		svcCtx.OptionTradingCalendarExceptionModel,
		contract,
		now,
	)
}

func IsContractTradingOpenWithModels(
	ctx context.Context,
	haltModel models.TOptionTradingHaltModel,
	calendarModel models.TOptionTradingCalendarModel,
	sessionModel models.TOptionTradingCalendarSessionModel,
	exceptionModel models.TOptionTradingCalendarExceptionModel,
	contract *models.TOptionContract,
	now int64,
) (*TradingCalendarDecision, error) {
	if contract == nil {
		return nil, errors.New("option contract is nil")
	}
	code, valid := NormalizeTradingCalendarCode(contract.TradingCalendarCode)
	if !valid {
		return nil, errors.New("option contract trading calendar code is invalid")
	}
	if _, err := haltModel.FindActiveByContract(
		ctx, contract.TenantId, contract.Id,
	); err == nil {
		return &TradingCalendarDecision{Reason: "ACTIVE_TRADING_HALT"}, nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	calendar, err := calendarModel.FindEffective(
		ctx, contract.TenantId, code, now,
	)
	if err != nil {
		return nil, err
	}
	sessions, err := sessionModel.FindByCalendar(
		ctx, calendar.TenantId, calendar.Id,
	)
	if err != nil {
		return nil, err
	}
	exceptions, err := exceptionModel.FindActiveAt(
		ctx, calendar.TenantId, calendar.Id, now,
	)
	if err != nil {
		return nil, err
	}
	decision, err := EvaluateTradingCalendarAt(calendar, sessions, exceptions, time.Unix(now, 0))
	if decision != nil {
		decision.CalendarId = calendar.Id
		decision.Version = calendar.Version
	}
	return decision, err
}

func EvaluateTradingCalendarAt(
	calendar *models.TOptionTradingCalendar,
	sessions []*models.TOptionTradingCalendarSession,
	exceptions []*models.TOptionTradingCalendarException,
	at time.Time,
) (*TradingCalendarDecision, error) {
	if calendar == nil {
		return nil, errors.New("trading calendar is nil")
	}
	location, err := time.LoadLocation(calendar.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load trading calendar timezone: %w", err)
	}
	hasOpenException := false
	for _, exception := range exceptions {
		if at.Unix() < exception.StartTime || at.Unix() >= exception.EndTime {
			continue
		}
		switch option.TradingCalendarExceptionType(exception.ExceptionType) {
		case option.TradingCalendarExceptionType_TRADING_CALENDAR_EXCEPTION_TYPE_CLOSED:
			return &TradingCalendarDecision{Reason: "CALENDAR_CLOSED_EXCEPTION"}, nil
		case option.TradingCalendarExceptionType_TRADING_CALENDAR_EXCEPTION_TYPE_OPEN:
			hasOpenException = true
		default:
			return nil, errors.New("invalid trading calendar exception type")
		}
	}
	if hasOpenException {
		return &TradingCalendarDecision{Open: true, Reason: "CALENDAR_OPEN_EXCEPTION"}, nil
	}
	local := at.In(location)
	currentDay := localDayStart(local, location)
	sourceDays := []time.Time{currentDay.AddDate(0, 0, -1), currentDay}
	for _, sourceDay := range sourceDays {
		for _, session := range sessions {
			if session.Weekday != int64(sourceDay.Weekday()) {
				continue
			}
			openAt := localSecond(sourceDay, session.OpenSecond, location)
			closeAt := localSecond(sourceDay, session.CloseSecond, location)
			if !at.Before(openAt) && at.Before(closeAt) {
				return &TradingCalendarDecision{Open: true, Reason: "WEEKLY_SESSION"}, nil
			}
		}
	}
	return &TradingCalendarDecision{Reason: "OUTSIDE_TRADING_SESSION"}, nil
}

func ValidateTradingCalendarDefinition(
	timezone string,
	effectiveFrom int64,
	effectiveUntil int64,
	sessions []*models.TOptionTradingCalendarSession,
	exceptions []*models.TOptionTradingCalendarException,
) error {
	if _, err := time.LoadLocation(strings.TrimSpace(timezone)); err != nil {
		return fmt.Errorf("invalid IANA timezone: %w", err)
	}
	if effectiveFrom < 0 || (effectiveUntil > 0 && effectiveUntil <= effectiveFrom) {
		return errors.New("invalid trading calendar effective period")
	}
	if len(sessions) == 0 {
		return errors.New("trading calendar requires at least one weekly session")
	}
	const weekSeconds = int64(7 * 86400)
	type interval struct{ start, end int64 }
	intervals := make([]interval, 0, len(sessions)*2)
	for _, session := range sessions {
		if session.Weekday < 0 || session.Weekday > 6 ||
			session.OpenSecond < 0 || session.OpenSecond >= 86400 ||
			session.CloseSecond <= session.OpenSecond || session.CloseSecond > 172800 {
			return errors.New("invalid trading calendar weekly session")
		}
		start := session.Weekday*86400 + session.OpenSecond
		end := session.Weekday*86400 + session.CloseSecond
		intervals = append(intervals, interval{start, end}, interval{start + weekSeconds, end + weekSeconds})
	}
	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i].start == intervals[j].start {
			return intervals[i].end < intervals[j].end
		}
		return intervals[i].start < intervals[j].start
	})
	for index := 1; index < len(intervals); index++ {
		if intervals[index].start < intervals[index-1].end {
			return errors.New("overlapping trading calendar weekly sessions")
		}
	}
	byType := map[int64][]interval{}
	for _, exception := range exceptions {
		if exception.ExceptionType != int64(option.TradingCalendarExceptionType_TRADING_CALENDAR_EXCEPTION_TYPE_CLOSED) &&
			exception.ExceptionType != int64(option.TradingCalendarExceptionType_TRADING_CALENDAR_EXCEPTION_TYPE_OPEN) {
			return errors.New("invalid trading calendar exception type")
		}
		if exception.StartTime <= 0 || exception.EndTime <= exception.StartTime ||
			strings.TrimSpace(exception.Reason) == "" {
			return errors.New("invalid trading calendar exception window")
		}
		if effectiveFrom > 0 && exception.StartTime < effectiveFrom {
			return errors.New("trading calendar exception starts before version")
		}
		if effectiveUntil > 0 && exception.EndTime > effectiveUntil {
			return errors.New("trading calendar exception ends after version")
		}
		byType[exception.ExceptionType] = append(
			byType[exception.ExceptionType],
			interval{exception.StartTime, exception.EndTime},
		)
	}
	for _, typed := range byType {
		sort.Slice(typed, func(i, j int) bool { return typed[i].start < typed[j].start })
		for index := 1; index < len(typed); index++ {
			if typed[index].start < typed[index-1].end {
				return errors.New("overlapping trading calendar exceptions of the same type")
			}
		}
	}
	return nil
}

func localDayStart(value time.Time, location *time.Location) time.Time {
	year, month, day := value.In(location).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func localSecond(day time.Time, second int64, location *time.Location) time.Time {
	date := day.AddDate(0, 0, int(second/86400))
	remainder := second % 86400
	year, month, dayOfMonth := date.Date()
	return time.Date(
		year, month, dayOfMonth,
		int(remainder/3600), int((remainder%3600)/60), int(remainder%60), 0, location,
	)
}
