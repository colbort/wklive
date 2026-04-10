package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type OptionContractPageFilter struct {
	TenantId         int64
	ContractCode     string
	UnderlyingSymbol string
	OptionType       int64
	Status           int64
	ListTimeStart    int64
	ListTimeEnd      int64
	ExpireTimeStart  int64
	ExpireTimeEnd    int64
}

type OptionContractModel interface {
	tOptionContractModel
	FindPage(ctx context.Context, filter OptionContractPageFilter, cursor int64, limit int64) ([]*TOptionContract, int64, error)
}

func (m *defaultTOptionContractModel) FindPage(ctx context.Context, filter OptionContractPageFilter, cursor int64, limit int64) ([]*TOptionContract, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("contract_code", filter.ContractCode)
	builder.EqString("underlying_symbol", filter.UnderlyingSymbol)
	builder.EqInt64("option_type", filter.OptionType)
	builder.EqInt64("status", filter.Status)
	builder.GteInt64("list_time", filter.ListTimeStart)
	builder.LteInt64("list_time", filter.ListTimeEnd)
	builder.GteInt64("expire_time", filter.ExpireTimeStart)
	builder.LteInt64("expire_time", filter.ExpireTimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionContractRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionContract
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
