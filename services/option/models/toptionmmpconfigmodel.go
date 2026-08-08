package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMmpConfigModel = (*customTOptionMmpConfigModel)(nil)

type (
	OptionMmpConfigPageFilter struct {
		TenantId   int64
		UserId     int64
		ContractId int64
		GroupCode  string
		Status     int64
	}

	// TOptionMmpConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMmpConfigModel.
	TOptionMmpConfigModel interface {
		tOptionMmpConfigModel
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionMmpConfig, error)
		FindForUpdate(ctx context.Context, tenantId, userId, contractId int64, groupCode string) (*TOptionMmpConfig, error)
		FindPage(ctx context.Context, filter OptionMmpConfigPageFilter, cursor, limit int64) ([]*TOptionMmpConfig, int64, error)
	}

	customTOptionMmpConfigModel struct {
		*defaultTOptionMmpConfigModel
	}
)

func (m *customTOptionMmpConfigModel) FindOneForUpdate(
	ctx context.Context, id int64,
) (*TOptionMmpConfig, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionMmpConfigRows, m.table)
	var item TOptionMmpConfig
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionMmpConfigModel) FindForUpdate(
	ctx context.Context, tenantId, userId, contractId int64, groupCode string,
) (*TOptionMmpConfig, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ? AND contract_id = ? AND group_code = ?
LIMIT 1 FOR UPDATE`, tOptionMmpConfigRows, m.table)
	var item TOptionMmpConfig
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, userId, contractId, groupCode); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionMmpConfigModel) FindPage(
	ctx context.Context, filter OptionMmpConfigPageFilter, cursor, limit int64,
) ([]*TOptionMmpConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("contract_id", filter.ContractId)
	builder.EqString("group_code", filter.GroupCode)
	builder.EqInt64("status", filter.Status)
	where := builder.Where()
	args := builder.Args()
	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	listSQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionMmpConfigRows, m.table, where)
	if cursor > 0 {
		listSQL += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	listSQL += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionMmpConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &items, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionMmpConfigModel returns a model for the database table.
func NewTOptionMmpConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMmpConfigModel {
	return &customTOptionMmpConfigModel{
		defaultTOptionMmpConfigModel: newTOptionMmpConfigModel(conn, c, opts...),
	}
}
