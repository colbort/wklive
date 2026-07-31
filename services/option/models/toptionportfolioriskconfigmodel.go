package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"
	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionPortfolioRiskConfigModel = (*customTOptionPortfolioRiskConfigModel)(nil)

type (
	OptionPortfolioRiskConfigPageFilter struct {
		TenantId   int64
		SettleCoin string
		Status     int64
	}

	// TOptionPortfolioRiskConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionPortfolioRiskConfigModel.
	TOptionPortfolioRiskConfigModel interface {
		tOptionPortfolioRiskConfigModel
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionPortfolioRiskConfig, error)
		FindLatestForUpdate(ctx context.Context, tenantId int64, settleCoin string) (*TOptionPortfolioRiskConfig, error)
		FindOpenEndedForUpdate(ctx context.Context, tenantId int64, settleCoin string) (*TOptionPortfolioRiskConfig, error)
		FindActive(ctx context.Context, tenantId int64, settleCoin string, now int64) (*TOptionPortfolioRiskConfig, error)
		FindPage(ctx context.Context, filter OptionPortfolioRiskConfigPageFilter, cursor, limit int64) ([]*TOptionPortfolioRiskConfig, int64, error)
	}

	customTOptionPortfolioRiskConfigModel struct {
		*defaultTOptionPortfolioRiskConfigModel
	}
)

func (m *defaultTOptionPortfolioRiskConfigModel) FindOneForUpdate(
	ctx context.Context,
	id int64,
) (*TOptionPortfolioRiskConfig, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE",
		tOptionPortfolioRiskConfigRows, m.table)
	var item TOptionPortfolioRiskConfig
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionPortfolioRiskConfigModel) FindLatestForUpdate(
	ctx context.Context,
	tenantId int64,
	settleCoin string,
) (*TOptionPortfolioRiskConfig, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id = ? AND settle_coin = ? ORDER BY version DESC LIMIT 1 FOR UPDATE",
		tOptionPortfolioRiskConfigRows, m.table,
	)
	var item TOptionPortfolioRiskConfig
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, settleCoin); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionPortfolioRiskConfigModel) FindOpenEndedForUpdate(
	ctx context.Context,
	tenantId int64,
	settleCoin string,
) (*TOptionPortfolioRiskConfig, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id = ? AND settle_coin = ? AND status IN (?, ?) "+
			"AND effective_until = 0 ORDER BY effective_from DESC, version DESC LIMIT 1 FOR UPDATE",
		tOptionPortfolioRiskConfigRows, m.table,
	)
	var item TOptionPortfolioRiskConfig
	if err := m.QueryRowNoCacheCtx(
		ctx, &item, query, tenantId, settleCoin,
		int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_APPROVED),
		int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_SUPERSEDED),
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionPortfolioRiskConfigModel) FindActive(
	ctx context.Context,
	tenantId int64,
	settleCoin string,
	now int64,
) (*TOptionPortfolioRiskConfig, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE tenant_id = ? AND settle_coin = ? AND status IN (?, ?) "+
			"AND effective_from <= ? AND (effective_until = 0 OR effective_until > ?) "+
			"ORDER BY effective_from DESC, version DESC LIMIT 1",
		tOptionPortfolioRiskConfigRows, m.table,
	)
	var item TOptionPortfolioRiskConfig
	if err := m.QueryRowNoCacheCtx(
		ctx, &item, query, tenantId, settleCoin,
		int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_APPROVED),
		int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_SUPERSEDED),
		now, now,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionPortfolioRiskConfigModel) FindPage(
	ctx context.Context,
	filter OptionPortfolioRiskConfigPageFilter,
	cursor, limit int64,
) ([]*TOptionPortfolioRiskConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("settle_coin", filter.SettleCoin)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(
		ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...,
	); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionPortfolioRiskConfigRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionPortfolioRiskConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionPortfolioRiskConfigModel returns a model for the database table.
func NewTOptionPortfolioRiskConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionPortfolioRiskConfigModel {
	return &customTOptionPortfolioRiskConfigModel{
		defaultTOptionPortfolioRiskConfigModel: newTOptionPortfolioRiskConfigModel(conn, c, opts...),
	}
}
