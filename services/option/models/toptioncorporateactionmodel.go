package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionCorporateActionModel = (*customTOptionCorporateActionModel)(nil)

type (
	OptionCorporateActionPageFilter struct {
		TenantId         int64
		UnderlyingSymbol string
		ActionType       int64
		Status           int64
	}

	// TOptionCorporateActionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionCorporateActionModel.
	TOptionCorporateActionModel interface {
		tOptionCorporateActionModel
		FindLatestVersionForUpdate(ctx context.Context, tenantId int64, externalEventRef string) (*TOptionCorporateAction, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionCorporateAction, error)
		FindDue(ctx context.Context, tenantId, now, limit int64) ([]*TOptionCorporateAction, error)
		FindPage(ctx context.Context, filter OptionCorporateActionPageFilter, cursor, limit int64) ([]*TOptionCorporateAction, int64, error)
	}

	customTOptionCorporateActionModel struct {
		*defaultTOptionCorporateActionModel
	}
)

func (m *customTOptionCorporateActionModel) FindLatestVersionForUpdate(
	ctx context.Context, tenantId int64, externalEventRef string,
) (*TOptionCorporateAction, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id=? AND external_event_ref=? ORDER BY version DESC LIMIT 1 FOR UPDATE",
		tOptionCorporateActionRows, m.table)
	var item TOptionCorporateAction
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, externalEventRef); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionCorporateActionModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionCorporateAction, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1 FOR UPDATE", tOptionCorporateActionRows, m.table)
	var item TOptionCorporateAction
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionCorporateActionModel) FindDue(
	ctx context.Context, tenantId, now, limit int64,
) ([]*TOptionCorporateAction, error) {
	limit = sqlutil.NormalizeLimit(limit)
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND status IN (?, ?, ?) AND effective_time<=?
ORDER BY effective_time, id LIMIT ?`, tOptionCorporateActionRows, m.table)
	var items []*TOptionCorporateAction
	err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, 2, 4, 7, now, limit)
	return items, err
}

func (m *customTOptionCorporateActionModel) FindPage(
	ctx context.Context, filter OptionCorporateActionPageFilter, cursor, limit int64,
) ([]*TOptionCorporateAction, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqString("underlying_symbol", filter.UnderlyingSymbol)
	builder.EqInt64("action_type", filter.ActionType)
	builder.EqInt64("status", filter.Status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total,
		fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionCorporateActionRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionCorporateAction
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionCorporateActionModel returns a model for the database table.
func NewTOptionCorporateActionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionCorporateActionModel {
	return &customTOptionCorporateActionModel{
		defaultTOptionCorporateActionModel: newTOptionCorporateActionModel(conn, c, opts...),
	}
}
