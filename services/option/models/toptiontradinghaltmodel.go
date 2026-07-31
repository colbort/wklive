package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradingHaltModel = (*customTOptionTradingHaltModel)(nil)

type (
	OptionTradingHaltPageFilter struct {
		TenantId   int64
		ContractId int64
		Status     int64
	}

	// TOptionTradingHaltModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradingHaltModel.
	TOptionTradingHaltModel interface {
		tOptionTradingHaltModel
		FindActiveByContract(ctx context.Context, tenantId, contractId int64) (*TOptionTradingHalt, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionTradingHalt, error)
		FindPage(ctx context.Context, filter OptionTradingHaltPageFilter, cursor, limit int64) ([]*TOptionTradingHalt, int64, error)
	}

	customTOptionTradingHaltModel struct {
		*defaultTOptionTradingHaltModel
	}
)

func (m *defaultTOptionTradingHaltModel) FindActiveByContract(
	ctx context.Context, tenantId, contractId int64,
) (*TOptionTradingHalt, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND active_key=? LIMIT 1`, tOptionTradingHaltRows, m.table)
	var item TOptionTradingHalt
	if err := m.QueryRowNoCacheCtx(
		ctx, &item, query, tenantId, fmt.Sprintf("CONTRACT:%d", contractId),
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionTradingHaltModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionTradingHalt, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1 FOR UPDATE",
		tOptionTradingHaltRows, m.table)
	var item TOptionTradingHalt
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionTradingHaltModel) FindPage(
	ctx context.Context, filter OptionTradingHaltPageFilter, cursor, limit int64,
) ([]*TOptionTradingHalt, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...,
	); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionTradingHaltRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionTradingHalt
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionTradingHaltModel returns a model for the database table.
func NewTOptionTradingHaltModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradingHaltModel {
	return &customTOptionTradingHaltModel{
		defaultTOptionTradingHaltModel: newTOptionTradingHaltModel(conn, c, opts...),
	}
}
