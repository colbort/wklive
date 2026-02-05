package models

import (
	"context"
	"fmt"
	"strings"
)

type UserRoleModel interface {
	sysUserRoleModel
	FindRoleIdsByUserId(ctx context.Context, uid int64) ([]int64, error)
	FindRoleIdsByUserIds(ctx context.Context, userIds []int64) (map[int64][]int64, error)
}

func (m *defaultSysUserRoleModel) FindRoleIdsByUserId(ctx context.Context, uid int64) ([]int64, error) {
	var ids []int64
	query := "select role_id from " + m.table + " where user_id = ?"
	err := m.conn.QueryRowsCtx(ctx, &ids, query, uid)
	return ids, err
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
	err := m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	mapping := make(map[int64][]int64)
	for _, r := range rows {
		mapping[r.UserId] = append(mapping[r.UserId], r.RoleId)
	}

	return mapping, nil
}
