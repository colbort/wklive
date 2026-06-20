package models

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
	"strings"
	"wklive/common/sqlutil"
)

var _ TItickProductModel = (*customTItickProductModel)(nil)

type (
	ItickProductPageFilter struct {
		CategoryType int32
		CategoryName string
		Market       string
		Keyword      string
		Enabled      int32
		AppVisible   int32
		Symbol       string
	}

	// TItickProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickProductModel.
	TItickProductModel interface {
		tItickProductModel
		FindPage(ctx context.Context, filter ItickProductPageFilter, cursor int64, limit int64) ([]*TItickProduct, int64, error)
		FindByIds(ctx context.Context, ids []int64) ([]*TItickProduct, error)
		Upsert(ctx context.Context, data *TItickProduct) (sql.Result, error)
	}

	customTItickProductModel struct {
		*defaultTItickProductModel
	}
)

// NewTItickProductModel returns a model for the database table.
func NewTItickProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickProductModel {
	return &customTItickProductModel{
		defaultTItickProductModel: newTItickProductModel(conn, c, opts...),
	}
}

func (m *defaultTItickProductModel) FindPage(ctx context.Context, filter ItickProductPageFilter, cursor int64, limit int64) ([]*TItickProduct, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("category_type", int64(filter.CategoryType))
	builder.EqString("category_name", filter.CategoryName)
	builder.EqString("market", filter.Market)
	builder.EqInt64("enabled", int64(filter.Enabled))
	builder.EqInt64("app_visible", int64(filter.AppVisible))
	if strings.TrimSpace(filter.Symbol) != "" {
		builder.LikeString("symbol", filter.Symbol)
	}

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
			tItickProductRows, m.table, where,
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
			tItickProductRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TItickProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTItickProductModel) FindByIds(ctx context.Context, ids []int64) ([]*TItickProduct, error) {
	if len(ids) == 0 {
		return []*TItickProduct{}, nil
	}

	builder := sqlutil.NewPageQueryBuilder()
	builder.InInt64("id", ids)

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		tItickProductRows,
		m.table,
		builder.Where(),
	)

	var list []*TItickProduct
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, builder.Args()...); err != nil {
		return nil, err
	}

	return list, nil
}

func (m *defaultTItickProductModel) Upsert(ctx context.Context, data *TItickProduct) (sql.Result, error) {
	tItickProductCategoryTypeMarketSymbolKey := fmt.Sprintf("%s%v:%v:%v",
		cacheTItickProductCategoryTypeMarketSymbolPrefix,
		data.CategoryType, data.Market, data.Symbol,
	)
	tItickProductIdKey := fmt.Sprintf("%s%v", cacheTItickProductIdPrefix, data.Id)

	feilds := strings.Join(stringx.Remove(tItickProductFieldNames, "`id`"), ",")

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
	}, tItickProductCategoryTypeMarketSymbolKey, tItickProductIdKey)

	return ret, err
}
