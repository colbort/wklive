package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TStakeReconciliationModel = (*customTStakeReconciliationModel)(nil)

type (
	// TStakeReconciliationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTStakeReconciliationModel.
	TStakeReconciliationModel interface {
		tStakeReconciliationModel
		Upsert(ctx context.Context, data *TStakeReconciliation) error
		FindAdminPage(ctx context.Context, filter StakeReconciliationPageFilter, cursor, limit int64) ([]*TStakeReconciliation, int64, error)
	}

	customTStakeReconciliationModel struct {
		*defaultTStakeReconciliationModel
	}
)

type StakeReconciliationPageFilter struct {
	TenantId           int64
	ReconciliationDate int64
	CoinSymbol         string
	Status             int64
}

// NewTStakeReconciliationModel returns a model for the database table.
func NewTStakeReconciliationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TStakeReconciliationModel {
	return &customTStakeReconciliationModel{
		defaultTStakeReconciliationModel: newTStakeReconciliationModel(conn, c, opts...),
	}
}

func (m *defaultTStakeReconciliationModel) Upsert(ctx context.Context, data *TStakeReconciliation) error {
	_, err := m.ExecNoCacheCtx(ctx, `INSERT INTO t_stake_reconciliation
		(tenant_id,reconciliation_date,coin_symbol,active_principal,product_staked,position_staked,asset_locked,
		 reward_log_amount,reward_platform_amount,fee_log_amount,fee_platform_amount,
		 product_diff,position_diff,lock_diff,reward_diff,fee_diff,status,detail,create_times,update_times)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
		 active_principal=VALUES(active_principal),product_staked=VALUES(product_staked),
		 position_staked=VALUES(position_staked),asset_locked=VALUES(asset_locked),
		 reward_log_amount=VALUES(reward_log_amount),reward_platform_amount=VALUES(reward_platform_amount),
		 fee_log_amount=VALUES(fee_log_amount),fee_platform_amount=VALUES(fee_platform_amount),
		 product_diff=VALUES(product_diff),position_diff=VALUES(position_diff),lock_diff=VALUES(lock_diff),
		 reward_diff=VALUES(reward_diff),fee_diff=VALUES(fee_diff),status=VALUES(status),detail=VALUES(detail),
		 update_times=VALUES(update_times)`,
		data.TenantId, data.ReconciliationDate, data.CoinSymbol, data.ActivePrincipal, data.ProductStaked,
		data.PositionStaked, data.AssetLocked, data.RewardLogAmount, data.RewardPlatformAmount,
		data.FeeLogAmount, data.FeePlatformAmount, data.ProductDiff, data.PositionDiff, data.LockDiff,
		data.RewardDiff, data.FeeDiff, data.Status, data.Detail, data.CreateTimes, data.UpdateTimes)
	return err
}

func (m *defaultTStakeReconciliationModel) FindAdminPage(ctx context.Context, filter StakeReconciliationPageFilter, cursor, limit int64) ([]*TStakeReconciliation, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	where := " WHERE 1=1"
	args := make([]any, 0, 5)
	if filter.TenantId > 0 {
		where += " AND tenant_id=?"
		args = append(args, filter.TenantId)
	}
	if filter.ReconciliationDate > 0 {
		where += " AND reconciliation_date=?"
		args = append(args, filter.ReconciliationDate)
	}
	if coin := strings.TrimSpace(filter.CoinSymbol); coin != "" {
		where += " AND coin_symbol=?"
		args = append(args, strings.ToUpper(coin))
	}
	if filter.Status > 0 {
		where += " AND status=?"
		args = append(args, filter.Status)
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, "SELECT COUNT(*) FROM "+m.table+where, args...); err != nil {
		return nil, 0, err
	}
	pageWhere := where
	pageArgs := append([]any(nil), args...)
	if cursor > 0 {
		pageWhere += " AND id<?"
		pageArgs = append(pageArgs, cursor)
	}
	pageArgs = append(pageArgs, limit)
	query := fmt.Sprintf("SELECT %s FROM %s%s ORDER BY id DESC LIMIT ?", tStakeReconciliationRows, m.table, pageWhere)
	var items []*TStakeReconciliation
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, pageArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
