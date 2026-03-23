package models

import (
	"context"
	"fmt"
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

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	where := "1=1"
	args := make([]any, 0, 4)

	if username != "" {
		where += " AND username LIKE ?"
		args = append(args, "%"+username+"%")
	}

	// 假设 success=0 表示全部，不筛选
	// 如果 0 是有效值，这里要改
	if success != 0 {
		where += " AND success = ?"
		args = append(args, success)
	}

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
