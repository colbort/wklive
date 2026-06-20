package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractLeverageConfigModel = (*customTContractLeverageConfigModel)(nil)

type (
	// TContractLeverageConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractLeverageConfigModel.
	TContractLeverageConfigModel interface {
		tContractLeverageConfigModel
		FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TContractLeverageConfig, int64, error)
	}

	customTContractLeverageConfigModel struct {
		*defaultTContractLeverageConfigModel
	}
)

// NewTContractLeverageConfigModel returns a model for the database table.
func NewTContractLeverageConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractLeverageConfigModel {
	return &customTContractLeverageConfigModel{
		defaultTContractLeverageConfigModel: newTContractLeverageConfigModel(conn, c, opts...),
	}
}

func (m *defaultTContractLeverageConfigModel) FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TContractLeverageConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)

	where := builder.Where()
	args := builder.Args()
	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s
            ORDER BY id DESC
            LIMIT ?`,
			tContractLeverageConfigRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tContractLeverageConfigRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TContractLeverageConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
