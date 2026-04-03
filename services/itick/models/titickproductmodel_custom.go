package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ItickProductModel interface {
	tItickProductModel
	FindPage(ctx context.Context, categoryType int32, market string, keyword string, enabled int32, appVisible int32, cursor int64, limit int64) ([]*TItickProduct, int64, error)
	Upsert(ctx context.Context, data *TItickProduct) (sql.Result, error)
}

func (m *defaultTItickProductModel) FindPage(ctx context.Context, categoryType int32, market string, keyword string, enabled int32, appVisible int32, cursor int64, limit int64) ([]*TItickProduct, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	where := "1=1"
	args := make([]any, 0, 2)

	if categoryType != 0 {
		where += " AND category_type = ?"
		args = append(args, categoryType)
	}

	if market != "" {
		where += " AND market = ?"
		args = append(args, market)
	}

	if enabled != 0 {
		where += " AND enabled = ?"
		args = append(args, enabled)
	}

	if appVisible != 0 {
		where += " AND app_visible = ?"
		args = append(args, appVisible)
	}

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

func (m *defaultTItickProductModel) Upsert(ctx context.Context, data *TItickProduct) (sql.Result, error) {
	tItickProductCategoryTypeMarketSymbolKey := fmt.Sprintf("%s%v:%v:%v",
		cacheTItickProductCategoryTypeMarketSymbolPrefix,
		data.CategoryType, data.Market, data.Symbol,
	)
	tItickProductIdKey := fmt.Sprintf("%s%v", cacheTItickProductIdPrefix, data.Id)

	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {

		query := fmt.Sprintf(`
			INSERT INTO %s (%s)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
				update_time = VALUES(update_time)
		`, m.table, tItickProductRowsExpectAutoSet)

		return conn.ExecCtx(ctx, query,
			data.CategoryType,
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
		)
	}, tItickProductCategoryTypeMarketSymbolKey, tItickProductIdKey)

	return ret, err
}
