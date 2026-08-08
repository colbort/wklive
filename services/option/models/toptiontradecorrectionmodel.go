package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradeCorrectionModel = (*customTOptionTradeCorrectionModel)(nil)

type (
	OptionTradeCorrectionPageFilter struct {
		TenantId   int64
		TradeId    int64
		ContractId int64
		Status     int64
	}

	// TOptionTradeCorrectionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradeCorrectionModel.
	TOptionTradeCorrectionModel interface {
		tOptionTradeCorrectionModel
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionTradeCorrection, error)
		FindActiveByTradeForUpdate(ctx context.Context, tenantId, tradeId int64) (*TOptionTradeCorrection, error)
		FindPage(ctx context.Context, filter OptionTradeCorrectionPageFilter, cursor, limit int64) ([]*TOptionTradeCorrection, int64, error)
	}

	customTOptionTradeCorrectionModel struct {
		*defaultTOptionTradeCorrectionModel
	}
)

func (m *customTOptionTradeCorrectionModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionTradeCorrection, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionTradeCorrectionRows, m.table)
	var item TOptionTradeCorrection
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionTradeCorrectionModel) FindActiveByTradeForUpdate(
	ctx context.Context, tenantId, tradeId int64,
) (*TOptionTradeCorrection, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND trade_id = ? AND status IN (?, ?, ?)
ORDER BY id DESC LIMIT 1 FOR UPDATE`, tOptionTradeCorrectionRows, m.table)
	var item TOptionTradeCorrection
	if err := m.QueryRowNoCacheCtx(
		ctx, &item, query, tenantId, tradeId,
		int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_PENDING_REVIEW),
		int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_EXECUTING),
		int64(option.TradeCorrectionStatus_TRADE_CORRECTION_STATUS_MANUAL_REVIEW),
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionTradeCorrectionModel) FindPage(
	ctx context.Context,
	filter OptionTradeCorrectionPageFilter,
	cursor, limit int64,
) ([]*TOptionTradeCorrection, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("trade_id", filter.TradeId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()

	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &total,
		fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where),
		args...,
	); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionTradeCorrectionRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionTradeCorrection
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionTradeCorrectionModel returns a model for the database table.
func NewTOptionTradeCorrectionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradeCorrectionModel {
	return &customTOptionTradeCorrectionModel{
		defaultTOptionTradeCorrectionModel: newTOptionTradeCorrectionModel(conn, c, opts...),
	}
}
