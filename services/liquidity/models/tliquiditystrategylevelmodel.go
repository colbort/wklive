package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TLiquidityStrategyLevelModel = (*customTLiquidityStrategyLevelModel)(nil)

type (
	LiquidityStrategyLevelInput struct {
		LevelNo                    int64
		BidSpreadBps, AskSpreadBps float64
		BidQty, AskQty             float64
		Enabled                    int64
	}
	// TLiquidityStrategyLevelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityStrategyLevelModel.
	TLiquidityStrategyLevelModel interface {
		tLiquidityStrategyLevelModel
		FindList(ctx context.Context, tenantID, configID int64, enabledOnly bool) ([]*TLiquidityStrategyLevel, error)
		CountEnabled(ctx context.Context, tenantID, configID int64) (int64, error)
		Replace(ctx context.Context, tenantID, configID int64, levels []LiquidityStrategyLevelInput, now int64) error
	}

	customTLiquidityStrategyLevelModel struct {
		*defaultTLiquidityStrategyLevelModel
	}
)

// NewTLiquidityStrategyLevelModel returns a model for the database table.
func NewTLiquidityStrategyLevelModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityStrategyLevelModel {
	return &customTLiquidityStrategyLevelModel{
		defaultTLiquidityStrategyLevelModel: newTLiquidityStrategyLevelModel(conn, c, opts...),
	}
}

func (m *customTLiquidityStrategyLevelModel) FindList(ctx context.Context, tenantID, configID int64, enabledOnly bool) ([]*TLiquidityStrategyLevel, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = ? AND config_id = ?", tLiquidityStrategyLevelRows, m.table)
	if enabledOnly {
		query += " AND enabled = 1"
	}
	query += " ORDER BY level_no ASC"
	var rows []*TLiquidityStrategyLevel
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, tenantID, configID); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customTLiquidityStrategyLevelModel) CountEnabled(ctx context.Context, tenantID, configID int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tenant_id = ? AND config_id = ? AND enabled = 1", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &count, query, tenantID, configID); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customTLiquidityStrategyLevelModel) Replace(ctx context.Context, tenantID, configID int64, levels []LiquidityStrategyLevelInput, now int64) error {
	oldRows, err := m.FindList(ctx, tenantID, configID, false)
	if err != nil {
		return err
	}
	err = m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := session.ExecCtx(ctx, fmt.Sprintf("DELETE FROM %s WHERE tenant_id = ? AND config_id = ?", m.table), tenantID, configID); err != nil {
			return err
		}
		for _, level := range levels {
			_, err := session.ExecCtx(ctx, fmt.Sprintf(`INSERT INTO %s
				(tenant_id, config_id, level_no, bid_spread_bps, ask_spread_bps, bid_qty, ask_qty, enabled, version, create_times, update_times)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`, m.table),
				tenantID, configID, level.LevelNo, level.BidSpreadBps, level.AskSpreadBps,
				level.BidQty, level.AskQty, level.Enabled, now, now)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(oldRows)*2)
	for _, row := range oldRows {
		keys = append(keys,
			fmt.Sprintf("%s%v", cacheTLiquidityStrategyLevelIdPrefix, row.Id),
			fmt.Sprintf("%s%v:%v:%v", cacheTLiquidityStrategyLevelTenantIdConfigIdLevelNoPrefix, row.TenantId, row.ConfigId, row.LevelNo),
		)
	}
	if len(keys) > 0 {
		return m.DelCacheCtx(ctx, keys...)
	}
	return nil
}
