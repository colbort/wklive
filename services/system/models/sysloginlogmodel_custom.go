package models

import (
	"context"
	"fmt"
	"wklive/common/sqlutil"
)

type LoginLogModel interface {
	sysLoginLogModel
	FindPage(ctx context.Context, username string, success int64, cursor, limit int64) ([]*SysLoginLog, int64, error)
}

func (m *defaultSysLoginLogModel) FindPage(
	ctx context.Context,
	username string,
	success int64,
	cursor, limit int64,
) ([]*SysLoginLog, int64, error) {

	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.LikeString("username", "%"+username+"%")
	builder.EqInt64("success", success)

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
