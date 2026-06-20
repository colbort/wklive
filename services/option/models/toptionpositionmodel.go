package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TOptionPositionModel = (*customTOptionPositionModel)(nil)

type (
	OptionPositionPageFilter struct {
		TenantId   int64
		UserId     int64
		AccountId  int64
		ContractId int64
		Side       int64
		Status     int64
		Statuses   []int64
	}

	// TOptionPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionPositionModel.
	TOptionPositionModel interface {
		tOptionPositionModel
		FindPage(ctx context.Context, filter OptionPositionPageFilter, cursor int64, limit int64) ([]*TOptionPosition, int64, error)
	}

	customTOptionPositionModel struct {
		*defaultTOptionPositionModel
	}
)

// NewTOptionPositionModel returns a model for the database table.
func NewTOptionPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionPositionModel {
	return &customTOptionPositionModel{
		defaultTOptionPositionModel: newTOptionPositionModel(conn, c, opts...),
	}
}

func (m *defaultTOptionPositionModel) FindPage(ctx context.Context, filter OptionPositionPageFilter, cursor int64, limit int64) ([]*TOptionPosition, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("side", filter.Side)
	builder.EqInt64("status", filter.Status)
	builder.InInt64("status", filter.Statuses)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionPositionRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
