package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"wklive/common/sqlutil"
	"wklive/proto/market"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TMarketQuoteModel = (*customTMarketQuoteModel)(nil)

type (
	MarketQuotePageFilter struct {
		Category string
		Symbol   string
	}

	// TMarketQuoteModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketQuoteModel.
	TMarketQuoteModel interface {
		tMarketQuoteModel
		Upsert(ctx context.Context, data *TMarketQuote) (sql.Result, error)
		FindPage(ctx context.Context, filter MarketQuotePageFilter, cursor int64, limit int64) ([]*TMarketQuote, int64, error)
		FindQuotes(ctx context.Context, data []*market.MarketSymbol) ([]*TMarketQuote, error)
	}

	customTMarketQuoteModel struct {
		*defaultTMarketQuoteModel
	}
)

// NewTMarketQuoteModel returns a model for the database table.
func NewTMarketQuoteModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TMarketQuoteModel {
	return &customTMarketQuoteModel{
		defaultTMarketQuoteModel: newTMarketQuoteModel(conn, c, opts...),
	}
}

func (m *defaultTMarketQuoteModel) Upsert(ctx context.Context, data *TMarketQuote) (sql.Result, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			last_price = VALUES(last_price),
			open_price = VALUES(open_price),
			high_price = VALUES(high_price),
			low_price = VALUES(low_price),
			prev_close_price = VALUES(prev_close_price),
			change_value = VALUES(change_value),
			change_rate = VALUES(change_rate),
			volume = VALUES(volume),
			turnover = VALUES(turnover),
			quote_ts = VALUES(quote_ts),
			trade_status = VALUES(trade_status),
			update_times = VALUES(update_times)
	`, m.table, tMarketQuoteRowsExpectAutoSet)

	marketQuoteMarketSymbolKey := fmt.Sprintf("%s%v:%v", cacheTMarketQuoteMarketSymbolPrefix, data.Market, data.Symbol)

	return m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query,
			data.Market,
			data.Symbol,
			data.LastPrice,
			data.OpenPrice,
			data.HighPrice,
			data.LowPrice,
			data.PrevClosePrice,
			data.ChangeValue,
			data.ChangeRate,
			data.Volume,
			data.Turnover,
			data.QuoteTs,
			data.TradeStatus,
			data.CreateTimes,
			data.UpdateTimes,
		)
	}, marketQuoteMarketSymbolKey)
}

func (m *defaultTMarketQuoteModel) FindPage(ctx context.Context, filter MarketQuotePageFilter, cursor int64, limit int64) ([]*TMarketQuote, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqString("category", filter.Category)
	builder.EqString("symbol", filter.Symbol)

	where := builder.Where()
	args := builder.Args()

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		// 第一页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s
			ORDER BY id DESC
			LIMIT ?`,
			tMarketQuoteRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		// 后续页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND id < ?
			ORDER BY id DESC
			LIMIT ?`,
			tMarketQuoteRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TMarketQuote
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTMarketQuoteModel) FindQuotes(ctx context.Context, data []*market.MarketSymbol) ([]*TMarketQuote, error) {
	if len(data) == 0 {
		return []*TMarketQuote{}, nil
	}

	list := make([]*TMarketQuote, 0)

	for _, item := range data {
		market := item.Market
		symbol := strings.TrimSpace(item.Symbol)
		if symbol == "" {
			continue
		}

		item, err := m.FindOneByMarketSymbol(ctx, market, symbol)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				continue
			}
			return nil, err
		}

		list = append(list, item)
	}

	return list, nil
}
