package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type OptionAccountPageFilter struct {
	TenantId   int64
	Uid        int64
	AccountId  int64
	MarginCoin string
	Status     int64
}

type OptionAccountModel interface {
	tOptionAccountModel
	FindPage(ctx context.Context, filter OptionAccountPageFilter, cursor int64, limit int64) ([]*TOptionAccount, int64, error)
}

func (m *defaultTOptionAccountModel) FindPage(ctx context.Context, filter OptionAccountPageFilter, cursor int64, limit int64) ([]*TOptionAccount, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("uid", filter.Uid)
	builder.EqInt64("account_id", filter.AccountId)
	builder.EqString("margin_coin", filter.MarginCoin)
	builder.EqInt64("status", filter.Status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionAccountRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionAccount
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
