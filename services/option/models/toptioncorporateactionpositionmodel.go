package models

import (
	"context"
	"fmt"

	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionCorporateActionPositionModel = (*customTOptionCorporateActionPositionModel)(nil)

type (
	// TOptionCorporateActionPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionCorporateActionPositionModel.
	TOptionCorporateActionPositionModel interface {
		tOptionCorporateActionPositionModel
		FindPage(ctx context.Context, tenantId, actionId, actionContractId, status, cursor, limit int64) ([]*TOptionCorporateActionPosition, int64, error)
	}

	customTOptionCorporateActionPositionModel struct {
		*defaultTOptionCorporateActionPositionModel
	}
)

func (m *customTOptionCorporateActionPositionModel) FindPage(
	ctx context.Context, tenantId, actionId, actionContractId, status, cursor, limit int64,
) ([]*TOptionCorporateActionPosition, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("action_id", actionId)
	builder.EqInt64("action_contract_id", actionContractId)
	builder.EqInt64("status", status)
	where, args := builder.Where(), builder.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total,
		fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	listArgs := append([]any{}, args...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tOptionCorporateActionPositionRows, m.table, where)
	if cursor > 0 {
		query += " AND id < ?"
		listArgs = append(listArgs, cursor)
	}
	query += " ORDER BY id DESC LIMIT ?"
	listArgs = append(listArgs, limit)
	var items []*TOptionCorporateActionPosition
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, listArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// NewTOptionCorporateActionPositionModel returns a model for the database table.
func NewTOptionCorporateActionPositionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionCorporateActionPositionModel {
	return &customTOptionCorporateActionPositionModel{
		defaultTOptionCorporateActionPositionModel: newTOptionCorporateActionPositionModel(conn, c, opts...),
	}
}
