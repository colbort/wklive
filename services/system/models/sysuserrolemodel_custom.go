package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserRoleModel interface {
	sysUserRoleModel
	FindRoleIdsByUserId(ctx context.Context, uid int64) ([]int64, error)
	FindRoleIdsByUserIds(ctx context.Context, userIds []int64) (map[int64][]int64, error)
	InsertCtx(ctx context.Context, session sqlx.Session, data *SysUserRole) (sql.Result, error)
	FindLoginUserPerms(ctx context.Context, uid int64) ([]string, error)
	FindByIds(ctx context.Context, userId int64, roleIds []int64) ([]int64, error)
}

func (m *defaultSysUserRoleModel) FindRoleIdsByUserId(ctx context.Context, uid int64) ([]int64, error) {
	var ids []int64
	query := "select role_id from " + m.table + " where user_id = ?"
	err := m.QueryRowsNoCacheCtx(ctx, &ids, query, uid)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (m *defaultSysUserRoleModel) FindRoleIdsByUserIds(
	ctx context.Context,
	userIds []int64,
) (map[int64][]int64, error) {

	if len(userIds) == 0 {
		return map[int64][]int64{}, nil
	}

	placeholders := make([]string, 0, len(userIds))
	args := make([]any, 0, len(userIds))
	for _, id := range userIds {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	query := fmt.Sprintf(
		"SELECT user_id, role_id FROM %s WHERE user_id IN (%s)",
		m.table,
		strings.Join(placeholders, ","),
	)

	type row struct {
		UserId int64
		RoleId int64
	}

	var rows []row

	err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	mapping := make(map[int64][]int64)
	for _, r := range rows {
		mapping[r.UserId] = append(mapping[r.UserId], r.RoleId)
	}

	return mapping, nil
}

func (m *defaultSysUserRoleModel) InsertCtx(ctx context.Context, session sqlx.Session, data *SysUserRole) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (`user_id`, `role_id`) values (?, ?)", m.table)
	ret, err := session.ExecCtx(ctx, query, data.UserId, data.RoleId)
	return ret, err
}

func (m *defaultSysUserRoleModel) FindLoginUserPerms(ctx context.Context, uid int64) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT m.perms
		FROM %s ur
		INNER JOIN sys_role_menu rm ON ur.role_id = rm.role_id
		INNER JOIN sys_menu m ON rm.menu_id = m.id
		WHERE ur.user_id = ? AND m.perms != ''
	`, m.table)

	var perms []string
	err := m.QueryRowsNoCacheCtx(ctx, &perms, query, uid)
	if err != nil {
		return nil, err
	}
	m.SetCacheCtx(ctx, "", perms)
	return perms, nil
}

func (m *defaultSysUserRoleModel) FindByIds(ctx context.Context, userId int64, roleIds []int64) ([]int64, error) {
	if len(roleIds) == 0 {
		return []int64{}, nil
	}

	placeholders := make([]string, 0, len(roleIds))
	args := make([]any, 0, len(roleIds)+1)
	for _, id := range roleIds {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, userId)

	query := fmt.Sprintf(
		"SELECT id FROM %s WHERE user_id = ? AND role_id IN (%s)",
		m.table,
		strings.Join(placeholders, ","),
	)

	var ids []int64
	err := m.QueryRowsNoCacheCtx(ctx, &ids, query, args...)
	if err != nil {
		return nil, err
	}
	return ids, nil
}
