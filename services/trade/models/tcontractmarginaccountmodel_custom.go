package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type ContractMarginAccountPageFilter struct {
	TenantId    int64
	UserId      int64
	MarketType  int64
	MarginAsset string
}

type ContractMarginAccountModel interface {
	tContractMarginAccountModel
	FindPage(ctx context.Context, filter ContractMarginAccountPageFilter, cursor int64, limit int64) ([]*TContractMarginAccount, int64, error)
	FindList(ctx context.Context, filter ContractMarginAccountPageFilter) ([]*TContractMarginAccount, error)
}

func (m *defaultTContractMarginAccountModel) FindPage(ctx context.Context, filter ContractMarginAccountPageFilter, cursor int64, limit int64) ([]*TContractMarginAccount, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqString("margin_asset", filter.MarginAsset)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractMarginAccountRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TContractMarginAccount
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *defaultTContractMarginAccountModel) FindList(ctx context.Context, filter ContractMarginAccountPageFilter) ([]*TContractMarginAccount, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("market_type", filter.MarketType)
	builder.EqString("margin_asset", filter.MarginAsset)

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY id DESC", tContractMarginAccountRows, m.table, builder.Where())
	var list []*TContractMarginAccount
	if err := m.QueryRowsNoCacheCtx(ctx, &list, sql, builder.Args()...); err != nil {
		return nil, err
	}
	return list, nil
}
