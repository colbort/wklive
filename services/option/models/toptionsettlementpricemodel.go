package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionSettlementPriceModel = (*customTOptionSettlementPriceModel)(nil)

type (
	OptionSettlementPricePageFilter struct {
		TenantId   int64
		ContractId int64
		Status     int64
	}

	// TOptionSettlementPriceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionSettlementPriceModel.
	TOptionSettlementPriceModel interface {
		tOptionSettlementPriceModel
		FindLatest(ctx context.Context, tenantId, contractId int64) (*TOptionSettlementPrice, error)
		FindLatestForUpdate(ctx context.Context, tenantId, contractId int64) (*TOptionSettlementPrice, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionSettlementPrice, error)
		FindLatestConfirmed(ctx context.Context, tenantId, contractId int64) (*TOptionSettlementPrice, error)
		FindPage(ctx context.Context, filter OptionSettlementPricePageFilter, cursor, limit int64) ([]*TOptionSettlementPrice, int64, error)
	}

	customTOptionSettlementPriceModel struct {
		*defaultTOptionSettlementPriceModel
	}
)

func (m *defaultTOptionSettlementPriceModel) FindLatest(
	ctx context.Context,
	tenantId, contractId int64,
) (*TOptionSettlementPrice, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id = ? AND contract_id = ? ORDER BY version DESC LIMIT 1",
		tOptionSettlementPriceRows, m.table,
	)
	var item TOptionSettlementPrice
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, contractId); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionSettlementPriceModel) FindLatestForUpdate(
	ctx context.Context,
	tenantId, contractId int64,
) (*TOptionSettlementPrice, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id = ? AND contract_id = ? ORDER BY version DESC LIMIT 1 FOR UPDATE",
		tOptionSettlementPriceRows, m.table,
	)
	var item TOptionSettlementPrice
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, contractId); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionSettlementPriceModel) FindOneForUpdate(
	ctx context.Context,
	id int64,
) (*TOptionSettlementPrice, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionSettlementPriceRows, m.table)
	var item TOptionSettlementPrice
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionSettlementPriceModel) FindLatestConfirmed(
	ctx context.Context,
	tenantId, contractId int64,
) (*TOptionSettlementPrice, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id = ? AND contract_id = ? AND status = ? ORDER BY version DESC LIMIT 1",
		tOptionSettlementPriceRows, m.table,
	)
	var item TOptionSettlementPrice
	if err := m.QueryRowNoCacheCtx(
		ctx, &item, query, tenantId, contractId,
		int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED),
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionSettlementPriceModel) FindPage(
	ctx context.Context,
	filter OptionSettlementPricePageFilter,
	cursor, limit int64,
) ([]*TOptionSettlementPrice, int64, error) {
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
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionSettlementPriceRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionSettlementPrice
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionSettlementPriceModel returns a model for the database table.
func NewTOptionSettlementPriceModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionSettlementPriceModel {
	return &customTOptionSettlementPriceModel{
		defaultTOptionSettlementPriceModel: newTOptionSettlementPriceModel(conn, c, opts...),
	}
}
