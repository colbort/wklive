package models

import (
	"context"
	"fmt"
	"time"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserModel interface {
	tUserModel
	FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TUser, int64, error)
	FindByInviteCode(ctx context.Context, inviteCode string) (*TUser, error)
	CountRecentNoRecharge(ctx context.Context, id int64) (int64, error)
	FindByUsername(ctx context.Context, tenantCode string, username string) (*TUser, error)
	FindByDeviceIdOrFingerprint(ctx context.Context, deviceId string, fingerprint string) (*TUser, error)
	FindByTenantIdUserId(ctx context.Context, tenantId int64, userId int64) (*TUser, error)
}

func (m *defaultTUserModel) FindPage(ctx context.Context, tenantId int64, cursor int64, limit int64) ([]*TUser, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)

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
			tUserRows, m.table, where,
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
			tUserRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TUser
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTUserModel) FindByInviteCode(ctx context.Context, inviteCode string) (*TUser, error) {
	var resp TUser

	query := fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE invite_code = ? 
		LIMIT 1
	`, tUserRows, m.table)

	err := m.QueryRowNoCacheCtx(ctx, &resp, query, inviteCode)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}

func (m *defaultTUserModel) CountRecentNoRecharge(ctx context.Context, id int64) (int64, error) {
	var count int64

	query := fmt.Sprintf(`
		SELECT COUNT(*) AS cnt
		FROM %s u
		WHERE u.create_times >= ?
		AND is_recharge = 0
		AND referrer_user_id = ?
	`, m.table)

	// 7 天前毫秒时间戳
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).UnixMilli()

	err := m.QueryRowNoCacheCtx(ctx, &count, query, sevenDaysAgo, id)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *defaultTUserModel) FindByUsername(ctx context.Context, tenantCode string, username string) (*TUser, error) {
	var resp TUser

	query := fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE tenant_code = ? 
		AND username = ? 
		LIMIT 1
	`, tUserRows, m.table)

	err := m.QueryRowNoCacheCtx(ctx, &resp, query, tenantCode, username)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}

func (m *defaultTUserModel) FindByDeviceIdOrFingerprint(ctx context.Context, deviceId string, fingerprint string) (*TUser, error) {
	var resp TUser

	query := fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE device_id = ? 
		OR fingerprint = ?
		ORDER BY id DESC
		LIMIT 1
	`, tUserRows, m.table)

	err := m.QueryRowNoCacheCtx(ctx, &resp, query, deviceId, fingerprint)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}

func (m *defaultTUserModel) FindByTenantIdUserId(ctx context.Context, tenantId int64, userId int64) (*TUser, error) {
	var resp TUser

	query := fmt.Sprintf(`
		SELECT %s 
		FROM %s 
		WHERE tenant_id = ? 
		AND user_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, tUserRows, m.table)

	err := m.QueryRowNoCacheCtx(ctx, &resp, query, tenantId, userId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &resp, nil
}
