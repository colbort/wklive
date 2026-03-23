package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserModel interface {
	sysUserModel
	FindPage(ctx context.Context, keyword string, status, cursor, limit int64) ([]*SysUser, int64, error)
	TransCtx(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error
	InsertCtx(ctx context.Context, session sqlx.Session, data *SysUser) (sql.Result, error)
}

func (m *defaultSysUserModel) FindPage(
	ctx context.Context,
	keyword string,
	status int64,
	cursor, limit int64,
) ([]*SysUser, int64, error) {

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// ---- WHERE 条件 ----
	where := "1=1"
	args := make([]any, 0, 4)

	if keyword != "" {
		where += " AND (username LIKE ? OR nickname LIKE ?)"
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	// 假设 status < 0 表示全部
	if status > 0 {
		where += " AND status = ?"
		args = append(args, status)
	}

	// ---- total ----
	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	// ---- list ----
	var listSql string
	listArgs := append([]any{}, args...)

	if cursor <= 0 {
		// 第一页
		listSql = fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ?",
			sysUserRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		// 后续页
		listSql = fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s AND id < ? ORDER BY id DESC LIMIT ?",
			sysUserRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*SysUser
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultSysUserModel) TransCtx(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error {
	return m.TransCtx(ctx, fn)
}

func (m *defaultSysUserModel) InsertCtx(ctx context.Context, session sqlx.Session, data *SysUser) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (`username`, `nickname`, `password`, `status`) values (?, ?, ?, ?)", m.table)
	ret, err := session.ExecCtx(ctx, query, data.Username, data.Nickname, data.Password, data.Status)
	return ret, err
}
