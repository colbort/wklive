package models

import (
	"context"
	"database/sql"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AssetLockModel interface {
	tAssetLockModel
	FindPage(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, bizType string, bizNo string, status int64, cursor int64, limit int64) ([]*TAssetLock, int64, error)
	// 解锁时更新锁仓记录
	UpdateUnlock(ctx context.Context, lockNo string, amount float64, updateTimes int64) (bool, error)
	// 扣减锁仓记录
	UpdateDeduct(ctx context.Context, lockNo string, amount float64, updateTimes int64) (bool, error)
}

func (m *defaultTAssetLockModel) FindPage(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, bizType string, bizNo string, status int64, cursor int64, limit int64) ([]*TAssetLock, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("user_id", userId)
	builder.EqInt64("wallet_type", walletType)
	builder.EqString("coin", coin)
	builder.EqString("biz_type", bizType)
	builder.EqString("biz_no", bizNo)
	builder.EqInt64("status", status)

	where := builder.Where()
	args := builder.Args()

	var total int64
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	var listSql string
	if cursor <= 0 {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s
            ORDER BY id DESC
            LIMIT ?`,
			tAssetLockRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tAssetLockRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TAssetLock
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 解锁时更新锁仓记录：unlock_amount += amount，remain_amount -= amount
// 当 remain_amount 为 0 时，状态改为 3（已解锁）；否则为 2（部分解锁）
func (m *defaultTAssetLockModel) UpdateUnlock(ctx context.Context, lockNo string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			unlock_amount = unlock_amount + ?,
			remain_amount = remain_amount - ?,
			status = CASE
				WHEN remain_amount - ? = 0 THEN 3
				ELSE 2
			END,
			update_times = ?
		WHERE lock_no = ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, amount, updateTimes, lockNo)
	})

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

// 扣减锁仓记录：remain_amount -= amount
// 当 remain_amount 为 0 时，状态改为 4（已关闭）；否则为 2（部分解锁）
func (m *defaultTAssetLockModel) UpdateDeduct(ctx context.Context, lockNo string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			remain_amount = remain_amount - ?,
			status = CASE
				WHEN remain_amount - ? = 0 THEN 4
				ELSE 2
			END,
			update_times = ?
		WHERE lock_no = ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, lockNo)
	})

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}
