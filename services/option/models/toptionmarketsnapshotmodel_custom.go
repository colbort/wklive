package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
)

type OptionMarketSnapshotPageFilter struct {
	TenantId      int64
	ContractId    int64
	SnapshotStart int64
	SnapshotEnd   int64
}

type OptionMarketSnapshotModel interface {
	tOptionMarketSnapshotModel
	FindPage(ctx context.Context, filter OptionMarketSnapshotPageFilter, cursor int64, limit int64) ([]*TOptionMarketSnapshot, int64, error)
}

func (m *defaultTOptionMarketSnapshotModel) FindPage(ctx context.Context, filter OptionMarketSnapshotPageFilter, cursor int64, limit int64) ([]*TOptionMarketSnapshot, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.GteInt64("snapshot_time", filter.SnapshotStart)
	builder.LteInt64("snapshot_time", filter.SnapshotEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionMarketSnapshotRows, m.table, where)
	if cursor > 0 {
		listSql += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSql += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TOptionMarketSnapshot
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
