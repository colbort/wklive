package models

import (
	"context"
	"fmt"
)

type LoginLogModel interface {
	sysLoginLogModel
	FindPage(ctx context.Context, username string, success int64, page, pageSize int64) ([]*SysLoginLog, int64, error)
}

func (m *defaultSysLoginLogModel) FindPage(
	ctx context.Context,
	username string,
	success int64,
	page, pageSize int64,
) ([]*SysLoginLog, int64, error) {

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	where := "1=1"
	args := make([]any, 0, 4)

	if username != "" {
		where += " AND username LIKE ?"
		args = append(args, "%"+username+"%")
	}

	if success != 0 {
		where += " AND success = ?"
		args = append(args, success)
	}

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	offset := (page - 1) * pageSize
	listSql := fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?`, sysLoginLogRows, m.table, where)

	listArgs := append(args, pageSize, offset)

	var list []*SysLoginLog
	if err := m.conn.QueryRowsCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
