package models

import (
	"context"
)

type OpLogModel interface {
	sysOpLogModel
	FindPage(ctx context.Context, username string, method string, path string, cursor, pageSize int64) ([]*SysOpLog, int64, error)
}

func (m *defaultSysOpLogModel) FindPage(
	ctx context.Context,
	username string,
	method string,
	path string,
	cursor, pageSize int64,
) ([]*SysOpLog, int64, error) {

	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// ---- WHERE 条件 ----
	where := "1=1"
	args := make([]any, 0, 4)

	if username != "" {
		where += " AND username LIKE ?"
		args = append(args, "%"+username+"%")
	}

	if method != "" {
		where += " AND method LIKE ?"
		args = append(args, "%"+method+"%")
	}

	if path != "" {
		where += " AND path LIKE ?"
		args = append(args, "%"+path+"%")
	}

	// ---- total ----
	var total int64
	countSql := "SELECT COUNT(1) FROM " + m.table + " WHERE " + where
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	listArgs := append([]any{}, args...)
	var listSql string

	if cursor <= 0 {
		// 第一页
		listSql = "SELECT " + sysOpLogRows +
			" FROM " + m.table +
			" WHERE " + where +
			" ORDER BY id DESC LIMIT ?"
		listArgs = append(listArgs, pageSize)
	} else {
		// 后续页
		listSql = "SELECT " + sysOpLogRows +
			" FROM " + m.table +
			" WHERE " + where +
			" AND id < ? ORDER BY id DESC LIMIT ?"
		listArgs = append(listArgs, cursor, pageSize)
	}

	var list []*SysOpLog
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
