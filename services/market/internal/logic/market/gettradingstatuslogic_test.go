package marketlogic

import (
	"context"
	"testing"
	"time"

	"wklive/proto/market"
	"wklive/services/market/internal/market/calendar"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"
)

type tradingCalendarModelStub struct {
	row      *models.TItickMarketCalendar
	sessions []*models.TItickMarketSession
	holiday  *models.TItickMarketHoliday
}

func (s tradingCalendarModelStub) Resolve(context.Context, string, string, string) (*models.TItickMarketCalendar, error) {
	return s.row, nil
}

func (s tradingCalendarModelStub) FindSessions(context.Context, int64) ([]*models.TItickMarketSession, error) {
	return s.sessions, nil
}

func (s tradingCalendarModelStub) FindHoliday(context.Context, int64, time.Time) (*models.TItickMarketHoliday, error) {
	return s.holiday, nil
}

type tradingProductCalendarStub struct {
	row *models.TItickMarketCalendar
}

func (s tradingProductCalendarStub) ResolveCalendar(context.Context, string, string, string) (*models.TItickMarketCalendar, error) {
	return s.row, nil
}

func TestGetTradingStatusUsesProductCalendar(t *testing.T) {
	row := &models.TItickMarketCalendar{
		Id: 9, CategoryCode: "stock", Market: "US", Exchange: "NASDAQ", Timezone: "UTC", WeekStart: 1,
	}
	model := tradingCalendarModelStub{
		row: row,
		sessions: []*models.TItickMarketSession{{
			Id: 1, SessionType: "regular", StartTime: "09:30", EndTime: "16:00", WeekdayMask: 62,
		}},
	}
	resolver := calendar.NewResolver(model, tradingProductCalendarStub{row: row}, time.Minute)
	logic := NewGetTradingStatusLogic(context.Background(), &svc.ServiceContext{MarketCalendarResolver: resolver})

	openResp, err := logic.GetTradingStatus(&market.GetTradingStatusReq{
		CategoryCode: "stock", Market: "US", Symbol: "AAPL",
		Timestamp: time.Date(2026, 8, 10, 10, 0, 0, 0, time.UTC).UnixMilli(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !openResp.GetData().GetIsOpen() || !openResp.GetData().GetProductSpecific() || openResp.GetData().GetCalendarId() != 9 || openResp.GetData().GetSessionType() != "regular" {
		t.Fatalf("unexpected open status: %+v", openResp.GetData())
	}

	closedResp, err := logic.GetTradingStatus(&market.GetTradingStatusReq{
		CategoryCode: "stock", Market: "US", Symbol: "AAPL",
		Timestamp: time.Date(2026, 8, 10, 18, 0, 0, 0, time.UTC).UnixMilli(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if closedResp.GetData().GetIsOpen() || closedResp.GetData().GetReason() != "market_closed" {
		t.Fatalf("unexpected closed status: %+v", closedResp.GetData())
	}
}

func TestGetTradingStatusCryptoFallback(t *testing.T) {
	resolver := calendar.NewResolver(tradingCalendarModelStub{}, nil, time.Minute)
	logic := NewGetTradingStatusLogic(context.Background(), &svc.ServiceContext{MarketCalendarResolver: resolver})
	resp, err := logic.GetTradingStatus(&market.GetTradingStatusReq{
		CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT",
		Timestamp: time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC).UnixMilli(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GetData().GetIsOpen() || resp.GetData().GetReason() != "crypto_24x7" || resp.GetData().GetSessionType() != "24x7" {
		t.Fatalf("unexpected crypto fallback: %+v", resp.GetData())
	}
}
