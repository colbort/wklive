package models

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MarketQuoteSource struct {
	CategoryCode string `db:"category_code"`
	Market       string `db:"market"`
	Symbol       string `db:"symbol"`
}

func (s *MarketQuoteSource) Source() string {
	if s == nil {
		return ""
	}
	categoryCode := strings.ToLower(strings.TrimSpace(s.CategoryCode))
	market := strings.ToUpper(strings.TrimSpace(s.Market))
	symbol := strings.ToUpper(strings.TrimSpace(s.Symbol))
	if categoryCode == "" || market == "" || symbol == "" {
		return ""
	}
	return categoryCode + ":" + market + ":" + symbol
}

type MarketQuoteSourceModel interface {
	FindEnabledTenantProduct(ctx context.Context, tenantID, categoryType int64, market, symbol string) (*MarketQuoteSource, error)
}

type marketQuoteSourceModel struct {
	conn sqlx.SqlConn
}

func NewMarketQuoteSourceModel(conn sqlx.SqlConn) MarketQuoteSourceModel {
	return &marketQuoteSourceModel{conn: conn}
}

func (m *marketQuoteSourceModel) FindEnabledTenantProduct(ctx context.Context, tenantID, categoryType int64, market, symbol string) (*MarketQuoteSource, error) {
	const query = `
		SELECT p.category_code, p.market, p.symbol
		FROM t_itick_tenant_product AS tp
		JOIN t_itick_product AS p ON p.id = tp.product_id
		WHERE tp.tenant_id = ?
		  AND p.category_type = ?
		  AND p.market = ?
		  AND p.symbol = ?
		  AND tp.enabled = 1
		  AND p.enabled = 1
		ORDER BY (tp.app_visible = 1) DESC, tp.sort ASC, p.sync_priority ASC, tp.id DESC
		LIMIT 1`

	var source MarketQuoteSource
	err := m.conn.QueryRowCtx(ctx, &source, query, tenantID, categoryType, strings.ToUpper(strings.TrimSpace(market)), strings.ToUpper(strings.TrimSpace(symbol)))
	if err != nil {
		return nil, err
	}
	return &source, nil
}
