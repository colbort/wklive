package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserModel interface {
	sysUserModel
	FindPage(ctx context.Context, keyword string, status int32, page, pageSize int64) ([]*SysUser, int64, error)
	TransCtx(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error
	InsertCtx(ctx context.Context, session sqlx.Session, data *SysUser) (sql.Result, error)
}

func (m *defaultSysUserModel) FindPage(
	ctx context.Context,
	keyword string,
	status int32,
	page, pageSize int64,
) ([]*SysUser, int64, error) {

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// ---- WHERE 条件 ----
	where := "1=1"
	args := make([]any, 0, 4)

	if keyword != "" {
		where += " AND (username LIKE ? OR nickname LIKE ?)"
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	if status != 0 {
		where += " AND status = ?"
		args = append(args, status)
	}

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	offset := (page - 1) * pageSize
	listSql := fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?`, sysUserRows, m.table, where)

	listArgs := append(args, pageSize, offset)

	var list []*SysUser
	if err := m.conn.QueryRowsCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultSysUserModel) TransCtx(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return fn(ctx, session)
	})
}

func (m *defaultSysUserModel) InsertCtx(ctx context.Context, session sqlx.Session, data *SysUser) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (`username`, `nickname`, `password`, `status`) values (?, ?, ?, ?)", m.table)
	ret, err := session.ExecCtx(ctx, query, data.Username, data.Nickname, data.Password, data.Status)
	return ret, err
}
