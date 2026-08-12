package marketlogic

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"wklive/common/helper"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTradingStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

const tradingStatusLogInterval = 5 * time.Minute

type tradingStatusLogEntry struct {
	state    string
	loggedAt time.Time
}

var tradingStatusLogs = struct {
	sync.Mutex
	entries map[string]tradingStatusLogEntry
}{entries: make(map[string]tradingStatusLogEntry)}

func NewGetTradingStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradingStatusLogic {
	return &GetTradingStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Resolves the product-specific or market-default calendar and reports
func (l *GetTradingStatusLogic) GetTradingStatus(in *market.GetTradingStatusReq) (*market.GetTradingStatusResp, error) {
	if in == nil {
		return nil, fmt.Errorf("invalid trading status query")
	}
	category := strings.ToLower(strings.TrimSpace(in.CategoryCode))
	marketCode := strings.ToUpper(strings.TrimSpace(in.Market))
	symbol := strings.ToUpper(strings.TrimSpace(in.Symbol))
	if category == "" || marketCode == "" || symbol == "" {
		return nil, fmt.Errorf("trading status product is required")
	}
	if l.svcCtx.MarketCalendarResolver == nil {
		return nil, fmt.Errorf("market calendar resolver is not configured")
	}
	timestamp := in.Timestamp
	if timestamp <= 0 {
		timestamp = time.Now().UnixMilli()
	}
	definition, err := l.svcCtx.MarketCalendarResolver.ResolveSymbolStrict(l.ctx, category, marketCode, symbol, "")
	if err != nil {
		return nil, fmt.Errorf("resolve market calendar: %w", err)
	}
	isOpen, sessionType, err := l.svcCtx.MarketCalendarResolver.EvaluateResolvedTradingSession(l.ctx, definition, category, timestamp)
	if err != nil {
		return nil, fmt.Errorf("evaluate market calendar: %w", err)
	}
	reason := "market_closed"
	if definition.ID == 0 {
		if isOpen {
			reason = "crypto_24x7"
		} else {
			reason = "calendar_not_found"
		}
	} else if isOpen {
		reason = "market_open"
	}
	logTradingStatusDecision(l.ctx, category, marketCode, symbol, definition.ID, definition.Timezone, isOpen, reason, sessionType)

	return &market.GetTradingStatusResp{
		Base: helper.OkResp(),
		Data: &market.GetTradingStatusData{
			IsOpen:          isOpen,
			CalendarId:      definition.ID,
			ProductSpecific: definition.ProductSpecific,
			Timezone:        definition.Timezone,
			Reason:          reason,
			SessionType:     sessionType,
		},
	}, nil
}

func logTradingStatusDecision(ctx context.Context, category, marketCode, symbol string, calendarID int64, timezone string, open bool, reason, sessionType string) {
	key := category + ":" + marketCode + ":" + symbol
	state := fmt.Sprintf("%t:%d:%s:%s", open, calendarID, reason, sessionType)
	now := time.Now()
	tradingStatusLogs.Lock()
	previous, found := tradingStatusLogs.entries[key]
	if found && previous.state == state && now.Sub(previous.loggedAt) < tradingStatusLogInterval {
		tradingStatusLogs.Unlock()
		return
	}
	tradingStatusLogs.entries[key] = tradingStatusLogEntry{state: state, loggedAt: now}
	tradingStatusLogs.Unlock()

	logger := logx.WithContext(ctx)
	if !open {
		logger.Errorf("[MARKET_TRADING_STATUS] product=%s calendar_id=%d timezone=%s open=false reason=%s session_type=%s", key, calendarID, timezone, reason, sessionType)
		return
	}
	logger.Infof("[MARKET_TRADING_STATUS] product=%s calendar_id=%d timezone=%s open=true reason=%s session_type=%s", key, calendarID, timezone, reason, sessionType)
}
