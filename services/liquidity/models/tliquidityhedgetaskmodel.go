package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityHedgeTaskModel = (*customTLiquidityHedgeTaskModel)(nil)

type (
	LiquidityHedgeTaskPageFilter struct {
		ConfigId, ProviderId, Status int64
		TimeStart, TimeEnd           int64
	}
	// TLiquidityHedgeTaskModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityHedgeTaskModel.
	TLiquidityHedgeTaskModel interface {
		tLiquidityHedgeTaskModel
		FindPage(ctx context.Context, filter LiquidityHedgeTaskPageFilter, cursor, limit int64) ([]*TLiquidityHedgeTask, int64, error)
	}

	customTLiquidityHedgeTaskModel struct {
		*defaultTLiquidityHedgeTaskModel
	}
)

// NewTLiquidityHedgeTaskModel returns a model for the database table.
func NewTLiquidityHedgeTaskModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityHedgeTaskModel {
	return &customTLiquidityHedgeTaskModel{
		defaultTLiquidityHedgeTaskModel: newTLiquidityHedgeTaskModel(conn, c, opts...),
	}
}

func (m *customTLiquidityHedgeTaskModel) FindPage(ctx context.Context, filter LiquidityHedgeTaskPageFilter, cursor, limit int64) ([]*TLiquidityHedgeTask, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := sqlutil.NewPageQueryBuilder()
	b.EqInt64("config_id", filter.ConfigId)
	b.EqInt64("provider_id", filter.ProviderId)
	b.EqInt64("status", filter.Status)
	b.GteInt64("create_times", filter.TimeStart)
	b.LteInt64("create_times", filter.TimeEnd)
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	queryArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tLiquidityHedgeTaskRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		queryArgs = append(queryArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	queryArgs = append(queryArgs, limit)
	var rows []*TLiquidityHedgeTask
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, queryArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
