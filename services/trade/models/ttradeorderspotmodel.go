package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TTradeOrderSpotModel = (*customTTradeOrderSpotModel)(nil)

type (
	// TTradeOrderSpotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeOrderSpotModel.
	TTradeOrderSpotModel interface {
		tTradeOrderSpotModel
		FindPage(ctx context.Context, cursor int64, limit int64) ([]*TTradeOrderSpot, int64, error)
	}

	customTTradeOrderSpotModel struct {
		*defaultTTradeOrderSpotModel
	}
)

// NewTTradeOrderSpotModel returns a model for the database table.
func NewTTradeOrderSpotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeOrderSpotModel {
	return &customTTradeOrderSpotModel{
		defaultTTradeOrderSpotModel: newTTradeOrderSpotModel(conn, c, opts...),
	}
}

func (m *defaultTTradeOrderSpotModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TTradeOrderSpot, int64, error) {
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
			tTradeOrderSpotRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tTradeOrderSpotRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TTradeOrderSpot
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
