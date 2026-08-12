package marketlogic

import (
	"context"
	"fmt"
	"strings"
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
	isOpen, err := l.svcCtx.MarketCalendarResolver.EvaluateResolvedTradingMinute(l.ctx, definition, category, timestamp)
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

	return &market.GetTradingStatusResp{
		Base: helper.OkResp(),
		Data: &market.GetTradingStatusData{
			IsOpen:          isOpen,
			CalendarId:      definition.ID,
			ProductSpecific: definition.ProductSpecific,
			Timezone:        definition.Timezone,
			Reason:          reason,
		},
	}, nil
}
