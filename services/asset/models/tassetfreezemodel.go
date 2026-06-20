package models

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TAssetFreezeModel = (*customTAssetFreezeModel)(nil)

type (
	AssetFreezePageFilter struct {
		TenantId   int64
		UserId     int64
		WalletType int64
		Coin       string
		BizType    string
		BizNo      string
		Status     int64
	}

	// TAssetFreezeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAssetFreezeModel.
	TAssetFreezeModel interface {
		tAssetFreezeModel
		FindPage(ctx context.Context, filter AssetFreezePageFilter, cursor int64, limit int64) ([]*TAssetFreeze, int64, error)
		// 解冻时更新冻结记录
		UpdateUnfreeze(ctx context.Context, freezeNo string, amount float64, updateTime int64) (bool, error)
		// 从冻结里扣减时更新冻结记录
		UpdateDeduct(ctx context.Context, freezeNo string, amount float64, updateTime int64) (bool, error)
	}

	customTAssetFreezeModel struct {
		*defaultTAssetFreezeModel
	}
)

// NewTAssetFreezeModel returns a model for the database table.
func NewTAssetFreezeModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAssetFreezeModel {
	return &customTAssetFreezeModel{
		defaultTAssetFreezeModel: newTAssetFreezeModel(conn, c, opts...),
	}
}

func (m *defaultTAssetFreezeModel) FindPage(ctx context.Context, filter AssetFreezePageFilter, cursor int64, limit int64) ([]*TAssetFreeze, int64, error) {
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
			tAssetFreezeRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
			tAssetFreezeRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TAssetFreeze
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *defaultTAssetFreezeModel) assetFreezeCacheKeys(ctx context.Context, freezeNo string) ([]string, error) {
	tAssetFreezeFreezeNoKey := fmt.Sprintf("%s%v", cacheTAssetFreezeFreezeNoPrefix, freezeNo)
	keys := []string{tAssetFreezeFreezeNoKey}

	var id int64
	query := fmt.Sprintf("SELECT `id` FROM %s WHERE `freeze_no` = ? LIMIT 1", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &id, query, freezeNo)
	switch err {
	case nil:
		tAssetFreezeIdKey := fmt.Sprintf("%s%v", cacheTAssetFreezeIdPrefix, id)
		keys = append(keys, tAssetFreezeIdKey)
		return keys, nil
	case sqlc.ErrNotFound, sql.ErrNoRows, ErrNotFound:
		return keys, nil
	default:
		return nil, err
	}
}

// 解冻时更新冻结记录：unfreeze_amount += amount，remain_amount -= amount
// 当 remain_amount 为 0 时，纯解冻为 3（已解冻），混合扣减/解冻为 5（已关闭）；否则为 2（部分释放）
func (m *defaultTAssetFreezeModel) UpdateUnfreeze(ctx context.Context, freezeNo string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			unfreeze_amount = unfreeze_amount + ?,
			remain_amount = remain_amount - ?,
			status = CASE
				WHEN remain_amount - ? <= 0 AND used_amount > 0 THEN 5
				WHEN remain_amount - ? <= 0 THEN 3
				ELSE 2
			END,
			update_times = ?
		WHERE freeze_no = ? AND status IN (1, 2) AND remain_amount >= ?
	`, m.table)

	cacheKeys, err := m.assetFreezeCacheKeys(ctx, freezeNo)
	if err != nil {
		return false, err
	}

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, amount, amount, updateTimes, freezeNo, amount)
	}, cacheKeys...)

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}

// 从冻结里扣减时更新冻结记录：used_amount += amount，remain_amount -= amount
// 当 remain_amount 为 0 时，纯扣减为 4（已扣完），混合扣减/解冻为 5（已关闭）；否则为 2（部分释放）
func (m *defaultTAssetFreezeModel) UpdateDeduct(ctx context.Context, freezeNo string, amount float64, updateTimes int64) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET 
			used_amount = used_amount + ?,
			remain_amount = remain_amount - ?,
			status = CASE
				WHEN remain_amount - ? <= 0 AND unfreeze_amount > 0 THEN 5
				WHEN remain_amount - ? <= 0 THEN 4
				ELSE 2
			END,
			update_times = ?
		WHERE freeze_no = ? AND status IN (1, 2) AND remain_amount >= ?
	`, m.table)

	cacheKeys, err := m.assetFreezeCacheKeys(ctx, freezeNo)
	if err != nil {
		return false, err
	}

	result, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, amount, amount, amount, amount, updateTimes, freezeNo, amount)
	}, cacheKeys...)

	if err != nil {
		return false, err
	}
	affected, _ := result.RowsAffected()
	return affected > 0, nil
}
