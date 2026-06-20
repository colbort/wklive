package models

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TAssetLockModel = (*customTAssetLockModel)(nil)

type (
	AssetLockPageFilter struct {
		TenantId   int64
		UserId     int64
		WalletType int64
		Coin       string
		BizType    string
		BizNo      string
		Status     int64
	}

	// TAssetLockModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetLockModel.
	TAssetLockModel interface {
		tAssetLockModel
		FindPage(ctx context.Context, filter AssetLockPageFilter, cursor int64, limit int64) ([]*TAssetLock, int64, error)
		// 解锁时更新锁仓记录
		UpdateUnlock(ctx context.Context, lockNo string, amount float64, updateTimes int64) (bool, error)
		// 扣减锁仓记录
		UpdateDeduct(ctx context.Context, lockNo string, amount float64, updateTimes int64) (bool, error)
	}

	customTAssetLockModel struct {
		*defaultTAssetLockModel
	}
)

// NewTAssetLockModel returns a model for the database table.
func NewTAssetLockModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetLockModel {
	return &customTAssetLockModel{
		defaultTAssetLockModel: newTAssetLockModel(conn, c, opts...),
	}
}

func (m *defaultTAssetLockModel) FindPage(ctx context.Context, filter AssetLockPageFilter, cursor int64, limit int64) ([]*TAssetLock, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("user_id", filter.UserId)
	builder.EqInt64("wallet_type", filter.WalletType)
	builder.EqString("coin", filter.Coin)
	builder.EqString("biz_type", filter.BizType)
	builder.EqString("biz_no", filter.BizNo)
	builder.EqInt64("status", filter.Status)

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
