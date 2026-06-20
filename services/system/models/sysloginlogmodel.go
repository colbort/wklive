package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ SysLoginLogModel = (*customSysLoginLogModel)(nil)

type (
	LoginLogPageFilter struct {
		Username string
		Success  int64
	}

	// SysLoginLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysLoginLogModel.
	SysLoginLogModel interface {
		sysLoginLogModel
		FindPage(ctx context.Context, filter LoginLogPageFilter, cursor int64, limit int64) ([]*SysLoginLog, int64, error)
	}

	customSysLoginLogModel struct {
		*defaultSysLoginLogModel
	}
)

// NewSysLoginLogModel returns a model for the database table.
func NewSysLoginLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysLoginLogModel {
	return &customSysLoginLogModel{
		defaultSysLoginLogModel: newSysLoginLogModel(conn, c, opts...),
	}
}

func (m *defaultSysLoginLogModel) FindPage(
	ctx context.Context,
	filter LoginLogPageFilter,
	cursor int64,
	limit int64,
) ([]*SysLoginLog, int64, error) {

	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.LikeString("username", filter.Username)
	builder.EqInt64("success", filter.Success)

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
			sysLoginLogRows, m.table, where,
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
			sysLoginLogRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*SysLoginLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
