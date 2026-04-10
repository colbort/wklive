package models

import (
	"context"
	"database/sql"
	"fmt"
	"wklive/common/sqlutil"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UserAssetModel interface {
	tUserAssetModel
	FindPage(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, status int64, cursor int64, limit int64) ([]*TUserAsset, int64, error)
	FindAll(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, status int64) ([]*TUserAsset, error)
	// 增加可用资产（充值等），如果不存在则先插入初始化记录
	AddAvailableAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, version int64, updateTimes int64) (int64, error)
	SubAvailableAmount(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error)
	// 冻结资产（下单冻结）
	FreezeAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error)
	// 解冻资产（撤单）
	UnfreezeAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error)
	// 从冻结里扣减（订单成交）
	DeductFromFrozen(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error)
	// 从锁仓里扣减（协议消耗）
	DeductLockedAmount(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error)
	// 锁仓（staking 参与）
	LockAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error)
	// 解锁
	UnlockAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error)
}

func (m *defaultTUserAssetModel) FindPage(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, status int64, cursor int64, limit int64) ([]*TUserAsset, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("user_id", userId)
	builder.EqInt64("wallet_type", walletType)
	builder.EqString("coin", coin)
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
			tUserAssetRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tUserAssetRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TUserAsset
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTUserAssetModel) FindAll(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, status int64) ([]*TUserAsset, error) {
	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", tenantId)
	builder.EqInt64("user_id", userId)
	builder.EqInt64("wallet_type", walletType)
	builder.EqString("coin", coin)
	builder.EqInt64("status", status)

	where := builder.Where()
	args := builder.Args()

	query := fmt.Sprintf(
		`SELECT %s
		FROM %s
		WHERE %s
		ORDER BY id DESC`,
		tUserAssetRows, m.table, where,
	)

	var list []*TUserAsset
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}

	return list, nil
}

// 增加可用资产（充值等）
func (m *defaultTUserAssetModel) AddAvailableAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, version int64, updateTimes int64) (int64, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			total_amount = total_amount + ?,
			available_amount = available_amount + ?,
			version = version + 1,
			update_times = ?
		WHERE tenant_id = ? AND user_id = ? AND wallet_type = ? AND coin = ? AND status = 1`,
		m.table)

	args := []any{amount, amount, updateTimes, tenantId, userId, walletType, coin}
	if version > 0 {
		args = append(args, version)
	}

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, args...)
	})

	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rows > 0 {
		return rows, nil
	}

	// 如果记录不存在，则插入新记录
	_, err = m.FindOneByTenantIdUserIdWalletTypeCoin(ctx, tenantId, userId, walletType, coin)
	if err == nil {
		return 0, nil
	}
	if err != sqlc.ErrNotFound {
		return 0, err
	}

	_, err = m.Insert(ctx, &TUserAsset{
		TenantId:        tenantId,
		UserId:          userId,
		WalletType:      walletType,
		Coin:            coin,
		TotalAmount:     amount,
		AvailableAmount: amount,
		FrozenAmount:    0,
		LockedAmount:    0,
		Status:          1,
		Version:         0,
		Remark:          "",
		CreateTimes:     updateTimes,
		UpdateTimes:     updateTimes,
	})
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (m *defaultTUserAssetModel) SubAvailableAmount(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			total_amount = total_amount - ?,
			available_amount = available_amount - ?,
			version = version + 1,
			update_times = ?
		WHERE tenant_id = ? AND user_id = ? AND wallet_type = ? AND coin = ? AND status = 1 AND available_amount >= ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, tenantId, userId, walletType, coin, amount)
	})
	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

func (m *defaultTUserAssetModel) DeductLockedAmount(ctx context.Context, tenantId int64, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			locked_amount = locked_amount - ?,
			total_amount = total_amount - ?,
			version = version + 1,
			update_times = ?
		WHERE tenant_id = ? AND user_id = ? AND wallet_type = ? AND coin = ? AND status = 1 AND locked_amount >= ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, tenantId, userId, walletType, coin, amount)
	})
	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

// 冻结资产（下单冻结）：可用 - amount，冻结 + amount
func (m *defaultTUserAssetModel) FreezeAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			available_amount = available_amount - ?,
			frozen_amount = frozen_amount + ?,
			version = version + 1,
			update_times = ?
		WHERE tenant_id = ? AND user_id = ? AND wallet_type = ? AND coin = ? 
			AND status = 1 AND available_amount >= ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, tenantId, userId, walletType, coin, amount)
	})

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

// 解冻资产（撤单）：可用 + amount，冻结 - amount
func (m *defaultTUserAssetModel) UnfreezeAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			available_amount = available_amount + ?,
			frozen_amount = frozen_amount - ?,
			version = version + 1,
			update_times = ?
		WHERE tenant_id = ? AND user_id = ? AND wallet_type = ? AND coin = ? 
			AND status = 1 AND frozen_amount >= ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, tenantId, userId, walletType, coin, amount)
	})

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

// 从冻结里扣减（订单成交）：冻结 - amount，总资产 - amount
func (m *defaultTUserAssetModel) DeductFromFrozen(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			frozen_amount = frozen_amount - ?,
			total_amount = total_amount - ?,
			version = version + 1,
			update_times = ?
		WHERE tenant_id = ? AND user_id = ? AND wallet_type = ? AND coin = ? 
			AND status = 1 AND frozen_amount >= ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, tenantId, userId, walletType, coin, amount)
	})

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

// 锁仓（staking 参与）：可用 - amount，锁定 + amount
func (m *defaultTUserAssetModel) LockAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			available_amount = available_amount - ?,
			locked_amount = locked_amount + ?,
			version = version + 1,
			update_times = ?
		WHERE tenant_id = ? AND user_id = ? AND wallet_type = ? AND coin = ? 
			AND status = 1 AND available_amount >= ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, tenantId, userId, walletType, coin, amount)
	})

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

// 解锁：可用 + amount，锁定 - amount
func (m *defaultTUserAssetModel) UnlockAmount(ctx context.Context, tenantId, userId int64, walletType int64, coin string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			available_amount = available_amount + ?,
			locked_amount = locked_amount - ?,
			version = version + 1,
			update_times = ?
		WHERE tenant_id = ? AND user_id = ? AND wallet_type = ? AND coin = ? 
			AND status = 1 AND locked_amount >= ?
	`, m.table)

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, updateTimes, tenantId, userId, walletType, coin, amount)
	})

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}
