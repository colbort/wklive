package models

import (
	"context"

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
	result, err := m.ExecNoCacheCtx(ctx, `INSERT IGNORE INTO t_option_market_snapshot_inbox
		(snapshot_id,tenant_id,contract_id,create_times) VALUES(?,?,?,?)`,
		snapshotID, tenantID, contractID, createTimes)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
