package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type OptionSettlementPageFilter struct {
	TenantId            int64
	ContractId          int64
	Status              int64
	SettlementTimeStart int64
	SettlementTimeEnd   int64
}

type OptionSettlementModel interface {
	tOptionSettlementModel
	FindPage(ctx context.Context, filter OptionSettlementPageFilter, cursor int64, limit int64) ([]*TOptionSettlement, int64, error)
}

func (m *defaultTOptionSettlementModel) FindPage(ctx context.Context, filter OptionSettlementPageFilter, cursor int64, limit int64) ([]*TOptionSettlement, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("status", filter.Status)
	builder.GteInt64("settlement_time", filter.SettlementTimeStart)
	builder.LteInt64("settlement_time", filter.SettlementTimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionSettlementRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionSettlement
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
