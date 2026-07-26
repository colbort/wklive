package models

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMarketSnapshotInboxModel = (*customTOptionMarketSnapshotInboxModel)(nil)

type (
	// TOptionMarketSnapshotInboxModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMarketSnapshotInboxModel.
	TOptionMarketSnapshotInboxModel interface {
		tOptionMarketSnapshotInboxModel
		Claim(ctx context.Context, snapshotID string, tenantID, contractID, createTimes int64) (bool, error)
		DeleteBefore(ctx context.Context, cutoff, limit int64) (int64, error)
	}

	customTOptionMarketSnapshotInboxModel struct {
		*defaultTOptionMarketSnapshotInboxModel
	}
)

// NewTOptionMarketSnapshotInboxModel returns a model for the database table.
func NewTOptionMarketSnapshotInboxModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMarketSnapshotInboxModel {
	return &customTOptionMarketSnapshotInboxModel{
		defaultTOptionMarketSnapshotInboxModel: newTOptionMarketSnapshotInboxModel(conn, c, opts...),
	}
}

func (m *defaultTOptionMarketSnapshotInboxModel) Claim(ctx context.Context, snapshotID string, tenantID, contractID, createTimes int64) (bool, error) {
	result, err := m.ExecNoCacheCtx(ctx, `INSERT INTO t_option_market_snapshot_inbox
		(snapshot_id,tenant_id,contract_id,create_times) VALUES(?,?,?,?)`,
		snapshotID, tenantID, contractID, createTimes)
	if err != nil {
		if isDuplicateKeyError(err) {
			return false, nil
		}
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (m *defaultTOptionMarketSnapshotInboxModel) DeleteBefore(ctx context.Context, cutoff, limit int64) (int64, error) {
	if limit <= 0 || limit > 10000 {
		limit = 5000
	}
	result, err := m.ExecNoCacheCtx(ctx, `DELETE FROM t_option_market_snapshot_inbox
		WHERE create_times<? ORDER BY id LIMIT ?`, cutoff, limit)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
