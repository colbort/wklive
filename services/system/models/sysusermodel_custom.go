package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	google2FASecretCachePrefix = "cache:google2FASecret:userId:"
)

type UserModel interface {
	sysUserModel
	FindPage(ctx context.Context, keyword string, status, cursor, limit int64) ([]*SysUser, int64, error)
	TransCtx(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error
	InsertCtx(ctx context.Context, session sqlx.Session, data *SysUser) (sql.Result, error)
	InsertGoogle2FASecret(ctx context.Context, userId int64, secret string) error
	GetGoogle2FASecret(ctx context.Context, userId int64) (string, error)
	DeleteGoogle2FASecret(ctx context.Context, userId int64) error
}

func (m *defaultSysUserModel) FindPage(
	ctx context.Context,
	keyword string,
	status int64,
	cursor, limit int64,
) ([]*SysUser, int64, error) {

	limit = sqlutil.NormalizeLimit(limit)

	// ---- WHERE 条件 ----
	builder := sqlutil.NewPageQueryBuilder()
	if keyword != "" {
		like := "%" + keyword + "%"
		builder.And("(username LIKE ? OR nickname LIKE ?)", like, like)
	}
	builder.EqInt64("status", status)

	where := builder.Where()
	args := builder.Args()

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

func (m *defaultSysUserModel) InsertGoogle2FASecret(ctx context.Context, userId int64, secret string) error {
	key := fmt.Sprintf("%s%v", google2FASecretCachePrefix, userId)
	return m.SetCacheWithExpireCtx(ctx, key, secret, 10*time.Minute)
}

func (m *defaultSysUserModel) GetGoogle2FASecret(ctx context.Context, userId int64) (string, error) {
	key := fmt.Sprintf("%s%v", google2FASecretCachePrefix, userId)
	var secret string
	err := m.GetCacheCtx(ctx, key, &secret)
	return secret, err
}

func (m *defaultSysUserModel) DeleteGoogle2FASecret(ctx context.Context, userId int64) error {
	key := fmt.Sprintf("%s%v", google2FASecretCachePrefix, userId)
	return m.DelCache(key)
}
