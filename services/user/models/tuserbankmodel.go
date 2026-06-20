package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TUserBankModel = (*customTUserBankModel)(nil)

type (
	UserBankPageFilter struct {
		TenantId int64
		UserId   int64
	}

	// TUserBankModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTUserBankModel.
	TUserBankModel interface {
		tUserBankModel
		FindPage(ctx context.Context, filter UserBankPageFilter, cursor int64, limit int64) ([]*TUserBank, int64, error)
	}

	customTUserBankModel struct {
		*defaultTUserBankModel
	}
)

// NewTUserBankModel returns a model for the database table.
func NewTUserBankModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TUserBankModel {
	return &customTUserBankModel{
		defaultTUserBankModel: newTUserBankModel(conn, c, opts...),
	}
}

func (m *defaultTUserBankModel) FindPage(ctx context.Context, filter UserBankPageFilter, cursor int64, limit int64) ([]*TUserBank, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)

	where := builder.Where()
	args := builder.Args()

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		// 第一页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s
			ORDER BY id DESC
			LIMIT ?`,
			tUserBankRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		// 后续页
		listSql = fmt.Sprintf(
			`SELECT %s
			FROM %s
			WHERE %s AND id < ?
			ORDER BY id DESC
			LIMIT ?`,
			tUserBankRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TUserBank
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
