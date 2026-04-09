package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AssetLockModel interface {
	tAssetLockModel
	FindPage(ctx context.Context, cursor int64, limit int64) ([]*TAssetLock, int64, error)
	// 解锁时更新锁仓记录
	UpdateUnlock(ctx context.Context, lockNo string, amount float64, updateTimes int64) (bool, error)
}

func (m *defaultTAssetLockModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TAssetLock, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	where := "1=1"
	args := make([]any, 0, 2)

	// ---- total ----
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
