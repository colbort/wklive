package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractPositionHistoryModel = (*customTContractPositionHistoryModel)(nil)

type (
	ContractPositionHistoryPageFilter struct {
		TenantId     int64
		UserId       int64
		SymbolId     int64
		ContractType int64
		PositionId   int64
		ActionType   int64
		TimeStart    int64
		TimeEnd      int64
	}

	// TContractPositionHistoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractPositionHistoryModel.
	TContractPositionHistoryModel interface {
		tContractPositionHistoryModel
		FindPage(ctx context.Context, filter ContractPositionHistoryPageFilter, cursor int64, limit int64) ([]*TContractPositionHistory, int64, error)
		CountByRefFillId(ctx context.Context, tenantID, fillID int64) (int64, error)
		FindByRefFillId(ctx context.Context, tenantID, fillID int64) ([]*TContractPositionHistory, error)
		FindLatestBySymbolAt(ctx context.Context, tenantID, symbolID, businessTime int64) ([]*TContractPositionHistory, error)
	}

	customTContractPositionHistoryModel struct {
		*defaultTContractPositionHistoryModel
	}
)

// NewTContractPositionHistoryModel returns a model for the database table.
func NewTContractPositionHistoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractPositionHistoryModel {
	return &customTContractPositionHistoryModel{
		defaultTContractPositionHistoryModel: newTContractPositionHistoryModel(conn, c, opts...),
	}
}

func (m *defaultTContractPositionHistoryModel) CountByRefFillId(ctx context.Context, tenantID, fillID int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id = ? AND ref_fill_id = ?", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &count, query, tenantID, fillID); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultTContractPositionHistoryModel) FindByRefFillId(ctx context.Context, tenantID, fillID int64) ([]*TContractPositionHistory, error) {
	var rows []*TContractPositionHistory
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND ref_fill_id=? ORDER BY id", tContractPositionHistoryRows, m.table)
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, tenantID, fillID); err != nil {
		return nil, err
	}
	return rows, nil
}

// FindLatestBySymbolAt returns the last recorded state for every position at
// the requested business time. It deliberately orders by business_time first,
// then id, so delayed event processing cannot move a fill into a later funding
// period merely because its history row was inserted later.
func (m *defaultTContractPositionHistoryModel) FindLatestBySymbolAt(ctx context.Context, tenantID, symbolID, businessTime int64) ([]*TContractPositionHistory, error) {
	query := fmt.Sprintf(`
SELECT %s
FROM %s AS h
WHERE h.tenant_id = ?
  AND h.symbol_id = ?
  AND h.business_time <= ?
  AND h.after_qty > 0
  AND NOT EXISTS (
    SELECT 1
    FROM %s AS newer
    WHERE newer.tenant_id = h.tenant_id
      AND newer.symbol_id = h.symbol_id
      AND newer.position_id = h.position_id
      AND newer.business_time <= ?
      AND (
        newer.business_time > h.business_time
        OR (newer.business_time = h.business_time AND newer.id > h.id)
      )
  )
ORDER BY h.position_id`, tContractPositionHistoryRows, m.table, m.table)
	var list []*TContractPositionHistory
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantID, symbolID, businessTime, businessTime); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultTContractPositionHistoryModel) FindPage(ctx context.Context, filter ContractPositionHistoryPageFilter, cursor int64, limit int64) ([]*TContractPositionHistory, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("contract_type", filter.ContractType)
	builder.EqInt64("position_id", filter.PositionId)
	builder.EqInt64("action_type", filter.ActionType)
	builder.GteInt64("create_times", filter.TimeStart)
	builder.LteInt64("create_times", filter.TimeEnd)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractPositionHistoryRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)

	var list []*TContractPositionHistory
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
