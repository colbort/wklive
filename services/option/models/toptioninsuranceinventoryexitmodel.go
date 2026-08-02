package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionInsuranceInventoryExitModel = (*customTOptionInsuranceInventoryExitModel)(nil)

type (
	OptionInsuranceInventoryExitPageFilter struct {
		TenantId   int64
		PositionId int64
		ContractId int64
		Status     int64
	}

	// TOptionInsuranceInventoryExitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionInsuranceInventoryExitModel.
	TOptionInsuranceInventoryExitModel interface {
		tOptionInsuranceInventoryExitModel
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionInsuranceInventoryExit, error)
		FindOpenByPositionForUpdate(ctx context.Context, tenantId, positionId int64) (*TOptionInsuranceInventoryExit, error)
		SumReservedQuantity(ctx context.Context, tenantId, contractId, utcDayStart int64) (decimal.Decimal, error)
		FindPage(ctx context.Context, filter OptionInsuranceInventoryExitPageFilter, cursor, limit int64) ([]*TOptionInsuranceInventoryExit, int64, error)
	}

	customTOptionInsuranceInventoryExitModel struct {
		*defaultTOptionInsuranceInventoryExitModel
	}
)

type insuranceInventoryExitQuantityAggregate struct {
	Total decimal.Decimal `db:"total"`
}

// SumReservedQuantity counts every open request as a conservative reservation,
// plus requests submitted during the current UTC day. Old open requests keep
// consuming budget until they are explicitly rejected.
func (m *defaultTOptionInsuranceInventoryExitModel) SumReservedQuantity(
	ctx context.Context, tenantId, contractId, utcDayStart int64,
) (decimal.Decimal, error) {
	query := fmt.Sprintf(`SELECT COALESCE(SUM(quantity), 0) AS total FROM %s
WHERE tenant_id=? AND contract_id=? AND (
  status IN (?,?) OR (status=? AND submitted_at>=?)
)`, m.table)
	var aggregate insuranceInventoryExitQuantityAggregate
	if err := m.QueryRowNoCacheCtx(ctx, &aggregate, query,
		tenantId, contractId, 1, 2, 4, utcDayStart); err != nil {
		return decimal.Zero, err
	}
	return aggregate.Total, nil
}

func (m *defaultTOptionInsuranceInventoryExitModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionInsuranceInventoryExit, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1 FOR UPDATE",
		tOptionInsuranceInventoryExitRows, m.table)
	var item TOptionInsuranceInventoryExit
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionInsuranceInventoryExitModel) FindOpenByPositionForUpdate(
	ctx context.Context, tenantId, positionId int64,
) (*TOptionInsuranceInventoryExit, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND position_id=? AND status IN (1,2)
ORDER BY id DESC LIMIT 1 FOR UPDATE`, tOptionInsuranceInventoryExitRows, m.table)
	var item TOptionInsuranceInventoryExit
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, positionId); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionInsuranceInventoryExitModel) FindPage(
	ctx context.Context,
	filter OptionInsuranceInventoryExitPageFilter,
	cursor, limit int64,
) ([]*TOptionInsuranceInventoryExit, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("position_id", filter.PositionId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total,
		fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionInsuranceInventoryExitRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionInsuranceInventoryExit
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionInsuranceInventoryExitModel returns a model for the database table.
func NewTOptionInsuranceInventoryExitModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionInsuranceInventoryExitModel {
	return &customTOptionInsuranceInventoryExitModel{
		defaultTOptionInsuranceInventoryExitModel: newTOptionInsuranceInventoryExitModel(conn, c, opts...),
	}
}
