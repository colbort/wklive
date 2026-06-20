package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TRiskUserTradeLimitModel = (*customTRiskUserTradeLimitModel)(nil)

type (
	// TRiskUserTradeLimitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTRiskUserTradeLimitModel.
	TRiskUserTradeLimitModel interface {
		tRiskUserTradeLimitModel
		FindPage(ctx context.Context, cursor int64, limit int64) ([]*TRiskUserTradeLimit, int64, error)
	}

	customTRiskUserTradeLimitModel struct {
		*defaultTRiskUserTradeLimitModel
	}
)

// NewTRiskUserTradeLimitModel returns a model for the database table.
func NewTRiskUserTradeLimitModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TRiskUserTradeLimitModel {
	return &customTRiskUserTradeLimitModel{
		defaultTRiskUserTradeLimitModel: newTRiskUserTradeLimitModel(conn, c, opts...),
	}
}

func (m *defaultTRiskUserTradeLimitModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TRiskUserTradeLimit, int64, error) {
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
			tRiskUserTradeLimitRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tRiskUserTradeLimitRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TRiskUserTradeLimit
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
