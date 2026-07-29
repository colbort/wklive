package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var _ TMarketProductModel = (*customTMarketProductModel)(nil)

type (
	MarketProductPageFilter struct {
		CategoryType int32
		CategoryName string
		Market       string
		Keyword      string
		Enabled      int32
		AppVisible   int32
		Symbol       string
	}

	// TMarketProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTMarketProductModel.
	TMarketProductModel interface {
		tMarketProductModel
		FindPage(ctx context.Context, filter MarketProductPageFilter, cursor int64, limit int64) ([]*TMarketProduct, int64, error)
		FindByIds(ctx context.Context, ids []int64) ([]*TMarketProduct, error)
		FindActivePage(ctx context.Context, cursor, limit int64) ([]*TMarketProduct, error)
		Upsert(ctx context.Context, data *TMarketProduct) (sql.Result, error)
	}

	customTMarketProductModel struct {
		*defaultTMarketProductModel
	}
)

// NewTMarketProductModel returns a model for the database table.
func NewTMarketProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TMarketProductModel {
	return &customTMarketProductModel{
		defaultTMarketProductModel: newTMarketProductModel(conn, c, opts...),
	}
}

// FindActivePage returns enabled products referenced by at least one enabled
// tenant. The EXISTS predicate naturally deduplicates products across tenants.
func (m *defaultTMarketProductModel) FindActivePage(ctx context.Context, cursor, limit int64) ([]*TMarketProduct, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf(`SELECT %s FROM %s AS p
		WHERE p.id > ? AND p.enabled = 1
		AND EXISTS (
			SELECT 1 FROM t_market_tenant_product AS tp
			WHERE tp.product_id = p.id AND tp.enabled = 1
		)
		ORDER BY p.id ASC LIMIT ?`, qualifyRows("p", tMarketProductRows), m.table)
	var list []*TMarketProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, cursor, limit); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTMarketProductModel) FindPage(ctx context.Context, filter MarketProductPageFilter, cursor int64, limit int64) ([]*TMarketProduct, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	queryLimit := limit + 1

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("category_type", int64(filter.CategoryType))
	builder.EqString("category_name", filter.CategoryName)
	builder.EqString("market", filter.Market)
	builder.EqInt64("enabled", int64(filter.Enabled))
	builder.EqInt64("app_visible", int64(filter.AppVisible))
	if strings.TrimSpace(filter.Symbol) != "" {
		builder.LikeString("symbol", filter.Symbol)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder.Or([]string{"name LIKE ?", "display_name LIKE ?", "code LIKE ?", "symbol LIKE ?"}, like, like, like, like)
	}

	where := builder.Where()
	args := builder.Args()

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
			tMarketProductRows, m.table, where,
		)
		listArgs = append(listArgs, queryLimit)
	} else {
		// 后续页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND id < ?
			ORDER BY id DESC
			LIMIT ?`,
			tMarketProductRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, queryLimit)
	}

	var list []*TMarketProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	// Cursor pagination only needs one extra row to determine hasNext. Returning
	// total=0 avoids an exact COUNT scan on every page request.
	return list, 0, nil
}

func (m *defaultTMarketProductModel) FindByIds(ctx context.Context, ids []int64) ([]*TMarketProduct, error) {
	if len(ids) == 0 {
		return []*TMarketProduct{}, nil
	}

	builder := sqlutil.NewPageQueryBuilder()
	builder.InInt64("id", ids)

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		tMarketProductRows,
		m.table,
		builder.Where(),
	)

	var list []*TMarketProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, builder.Args()...); err != nil {
		return nil, err
	}

	return list, nil
}

func (m *defaultTMarketProductModel) Upsert(ctx context.Context, data *TMarketProduct) (sql.Result, error) {
	tMarketProductCategoryTypeMarketSymbolKey := fmt.Sprintf("%s%v:%v:%v",
		cacheTMarketProductCategoryTypeMarketSymbolPrefix,
		data.CategoryType, data.Market, data.Symbol,
	)
	tMarketProductIdKey := fmt.Sprintf("%s%v", cacheTMarketProductIdPrefix, data.Id)

	feilds := strings.Join(stringx.Remove(tMarketProductFieldNames, "`id`"), ",")

	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf(`
            INSERT INTO %s (%s)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON DUPLICATE KEY UPDATE
                code = VALUES(code),
                name = VALUES(name),
                display_name = VALUES(display_name),
                exchange = VALUES(exchange),
                sector = VALUES(sector),
                lug = VALUES(lug),
                base_coin = VALUES(base_coin),
                quote_coin = VALUES(quote_coin),
                enabled = VALUES(enabled),
                app_visible = VALUES(app_visible),
                sort = VALUES(sort),
                icon = VALUES(icon),
                remark = VALUES(remark),
                update_times = VALUES(update_times)
        `, m.table, feilds)

		return conn.ExecCtx(ctx, query,
			data.CategoryType,
			data.CategoryName,
			data.CategoryCode,
			data.Market,
			data.Symbol,
			data.Code,
			data.Name,
			data.DisplayName,
			data.Exchange,
			data.Sector,
			data.Lug,
			data.BaseCoin,
			data.QuoteCoin,
			data.Enabled,
			data.AppVisible,
			data.Sort,
			data.Icon,
			data.Remark,
			data.CreateTimes,
			data.UpdateTimes,
		)
	}, tMarketProductCategoryTypeMarketSymbolKey, tMarketProductIdKey)

	return ret, err
}
