package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionPhysicalDeliveryUnitModel = (*customTOptionPhysicalDeliveryUnitModel)(nil)

type (
	OptionPhysicalDeliveryUnitPageFilter struct {
		TenantId   int64
		ContractId int64
		BatchId    int64
		UserId     int64
		Status     int64
	}

	// TOptionPhysicalDeliveryUnitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionPhysicalDeliveryUnitModel.
	TOptionPhysicalDeliveryUnitModel interface {
		tOptionPhysicalDeliveryUnitModel
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionPhysicalDeliveryUnit, error)
		FindByBatch(ctx context.Context, tenantId, batchId int64) ([]*TOptionPhysicalDeliveryUnit, error)
		FindExceptionByBatch(ctx context.Context, tenantId, batchId int64) (*TOptionPhysicalDeliveryUnit, error)
		FindExpiredCure(ctx context.Context, tenantId, now, limit int64) ([]*TOptionPhysicalDeliveryUnit, error)
		FindPage(ctx context.Context, filter OptionPhysicalDeliveryUnitPageFilter, cursor, limit int64) ([]*TOptionPhysicalDeliveryUnit, int64, error)
	}

	customTOptionPhysicalDeliveryUnitModel struct {
		*defaultTOptionPhysicalDeliveryUnitModel
	}
)

func (m *customTOptionPhysicalDeliveryUnitModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionPhysicalDeliveryUnit, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE",
		tOptionPhysicalDeliveryUnitRows, m.table)
	var item TOptionPhysicalDeliveryUnit
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionPhysicalDeliveryUnitModel) FindByBatch(
	ctx context.Context, tenantId, batchId int64,
) ([]*TOptionPhysicalDeliveryUnit, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND batch_id = ? ORDER BY id`, tOptionPhysicalDeliveryUnitRows, m.table)
	var items []*TOptionPhysicalDeliveryUnit
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, batchId); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *customTOptionPhysicalDeliveryUnitModel) FindExceptionByBatch(
	ctx context.Context, tenantId, batchId int64,
) (*TOptionPhysicalDeliveryUnit, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND batch_id = ? AND status IN (?, ?)
ORDER BY id LIMIT 1`, tOptionPhysicalDeliveryUnitRows, m.table)
	var item TOptionPhysicalDeliveryUnit
	if err := m.QueryRowNoCacheCtx(ctx, &item, query,
		tenantId, batchId,
		int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_MANUAL_REVIEW),
		int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_DEFAULTED),
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionPhysicalDeliveryUnitModel) FindExpiredCure(
	ctx context.Context, tenantId, now, limit int64,
) ([]*TOptionPhysicalDeliveryUnit, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE cure_deadline <= ? AND manual_retry_count = 0 AND status IN (?, ?)`, tOptionPhysicalDeliveryUnitRows, m.table)
	args := []any{
		now,
		int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_CURE_REQUIRED),
		int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_ASSET_PROCESSING),
	}
	if tenantId > 0 {
		query += " AND tenant_id = ?"
		args = append(args, tenantId)
	}
	query += " ORDER BY cure_deadline,id LIMIT ?"
	args = append(args, limit)
	var items []*TOptionPhysicalDeliveryUnit
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *customTOptionPhysicalDeliveryUnitModel) FindPage(
	ctx context.Context, filter OptionPhysicalDeliveryUnitPageFilter, cursor, limit int64,
) ([]*TOptionPhysicalDeliveryUnit, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("batch_id", filter.BatchId)
	builder.EqInt64("status", filter.Status)
	where := builder.Where()
	args := builder.Args()
	if filter.UserId > 0 {
		where += " AND (long_user_id = ? OR short_user_id = ?)"
		args = append(args, filter.UserId, filter.UserId)
	}
	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionPhysicalDeliveryUnitRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionPhysicalDeliveryUnit
	if err := m.QueryRowsNoCacheCtx(ctx, &items, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionPhysicalDeliveryUnitModel returns a model for the database table.
func NewTOptionPhysicalDeliveryUnitModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionPhysicalDeliveryUnitModel {
	return &customTOptionPhysicalDeliveryUnitModel{
		defaultTOptionPhysicalDeliveryUnitModel: newTOptionPhysicalDeliveryUnitModel(conn, c, opts...),
	}
}
