package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TRiskUserSymbolLimitModel = (*customTRiskUserSymbolLimitModel)(nil)

type (
	// TRiskUserSymbolLimitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRiskUserSymbolLimitModel.
	TRiskUserSymbolLimitModel interface {
		tRiskUserSymbolLimitModel
		FindPage(ctx context.Context, cursor int64, limit int64) ([]*TRiskUserSymbolLimit, int64, error)
	}

	customTRiskUserSymbolLimitModel struct {
		*defaultTRiskUserSymbolLimitModel
	}
)

// NewTRiskUserSymbolLimitModel returns a model for the database table.
func NewTRiskUserSymbolLimitModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRiskUserSymbolLimitModel {
	return &customTRiskUserSymbolLimitModel{
		defaultTRiskUserSymbolLimitModel: newTRiskUserSymbolLimitModel(conn, c, opts...),
	}
}

func (m *defaultTRiskUserSymbolLimitModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TRiskUserSymbolLimit, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	where := builder.Where()
	args := builder.Args()

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s
            ORDER BY id DESC
            LIMIT ?`,
			tRiskUserSymbolLimitRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tRiskUserSymbolLimitRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TRiskUserSymbolLimit
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
