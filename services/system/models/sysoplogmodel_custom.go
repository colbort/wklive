package models

import "context"

type OpLogModel interface {
	sysOpLogModel
	FindPage(ctx context.Context, username string, method string, path string, page, pageSize int64) ([]*SysOpLog, int64, error)
}

func (m *defaultSysOpLogModel) FindPage(
	ctx context.Context,
	username string,
	method string,
	path string,
	page, pageSize int64,
) ([]*SysOpLog, int64, error) {

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
	if err := m.conn.QueryRowCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	offset := (page - 1) * pageSize
	listSql := "SELECT " + sysOpLogRows + " FROM " + m.table + " WHERE " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"

	listArgs := append(args, pageSize, offset)

	var list []*SysOpLog
	if err := m.conn.QueryRowsCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
