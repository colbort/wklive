package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"wklive/common/sqlutil"
)

var _ TTradeSymbolLeverageConfigModel = (*customTTradeSymbolLeverageConfigModel)(nil)

type (
	TradeSymbolLeverageConfigPageFilter struct {
		TenantId   int64
		SymbolId   int64
		MarginMode int64
		Enabled    int64
	}

	TTradeSymbolLeverageConfigModel interface {
		tTradeSymbolLeverageConfigModel
		FindPage(ctx context.Context, filter TradeSymbolLeverageConfigPageFilter, cursor int64, limit int64) ([]*TTradeSymbolLeverageConfig, int64, error)
		SyncGroup(ctx context.Context, tenantId, symbolId, marginMode int64, leverageValues []int64, defaultLeverage, enabled, sort int64, remark string, now int64) error
		DisableGroup(ctx context.Context, tenantId, symbolId, marginMode, now int64) error
	}

	customTTradeSymbolLeverageConfigModel struct {
		*defaultTTradeSymbolLeverageConfigModel
	}
)

func (m *customTTradeSymbolLeverageConfigModel) DisableGroup(ctx context.Context, tenantId, symbolId, marginMode, now int64) error {
	var rows []*TTradeSymbolLeverageConfig
	var defaultIds []int64
	err := m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := session.QueryRowsCtx(ctx, &rows, "select "+tTradeSymbolLeverageConfigRows+" from `t_trade_symbol_leverage_config` where `tenant_id` = ? and `symbol_id` = ? and `margin_mode` = ?", tenantId, symbolId, marginMode); err != nil {
			return err
		}
		if err := session.QueryRowsCtx(ctx, &defaultIds, "select `id` from `t_trade_symbol_leverage_default` where `tenant_id` = ? and `symbol_id` = ? and `margin_mode` = ?", tenantId, symbolId, marginMode); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx, "update `t_trade_symbol_leverage_config` set `enabled` = 2, `update_times` = ? where `tenant_id` = ? and `symbol_id` = ? and `margin_mode` = ?", now, tenantId, symbolId, marginMode); err != nil {
			return err
		}
		_, err := session.ExecCtx(ctx, "delete from `t_trade_symbol_leverage_default` where `tenant_id` = ? and `symbol_id` = ? and `margin_mode` = ?", tenantId, symbolId, marginMode)
		return err
	})
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(rows)+len(defaultIds)+2)
	for _, row := range rows {
		keys = append(keys, fmt.Sprintf("%s%v", cacheTTradeSymbolLeverageConfigIdPrefix, row.Id))
	}
	keys = append(keys, fmt.Sprintf("%s%v:%v:%v", cacheTTradeSymbolLeverageConfigTenantIdSymbolIdMarginModePrefix, tenantId, symbolId, marginMode))
	for _, id := range defaultIds {
		keys = append(keys, fmt.Sprintf("%s%v", cacheTTradeSymbolLeverageDefaultIdPrefix, id))
	}
	keys = append(keys, fmt.Sprintf("%s%v:%v:%v", cacheTTradeSymbolLeverageDefaultTenantIdSymbolIdMarginModePrefix, tenantId, symbolId, marginMode))
	return m.DelCacheCtx(ctx, keys...)
}

func (m *customTTradeSymbolLeverageConfigModel) SyncGroup(ctx context.Context, tenantId, symbolId, marginMode int64, leverageValues []int64, defaultLeverage, enabled, sort int64, remark string, now int64) error {
	var oldRows []*TTradeSymbolLeverageConfig
	var oldDefaultIds []int64
	encodedValues, err := json.Marshal(leverageValues)
	if err != nil {
		return err
	}
	err = m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := session.QueryRowsCtx(ctx, &oldRows, "select "+tTradeSymbolLeverageConfigRows+" from `t_trade_symbol_leverage_config` where `tenant_id` = ? and `symbol_id` = ? and `margin_mode` = ?", tenantId, symbolId, marginMode); err != nil {
			return err
		}
		if err := session.QueryRowsCtx(ctx, &oldDefaultIds, "select `id` from `t_trade_symbol_leverage_default` where `tenant_id` = ? and `symbol_id` = ? and `margin_mode` = ?", tenantId, symbolId, marginMode); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx, "insert into `t_trade_symbol_leverage_config` (`tenant_id`, `symbol_id`, `margin_mode`, `leverage_values`, `enabled`, `sort`, `remark`, `create_times`, `update_times`) values (?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `leverage_values` = values(`leverage_values`), `enabled` = values(`enabled`), `sort` = values(`sort`), `remark` = values(`remark`), `update_times` = values(`update_times`)", tenantId, symbolId, marginMode, string(encodedValues), enabled, sort, remark, now, now); err != nil {
			return err
		}
		if enabled == 1 {
			_, err := session.ExecCtx(ctx, "insert into `t_trade_symbol_leverage_default` (`tenant_id`, `symbol_id`, `margin_mode`, `leverage`, `create_times`, `update_times`) values (?, ?, ?, ?, ?, ?) on duplicate key update `leverage` = values(`leverage`), `update_times` = values(`update_times`)", tenantId, symbolId, marginMode, defaultLeverage, now, now)
			return err
		}
		_, err := session.ExecCtx(ctx, "delete from `t_trade_symbol_leverage_default` where `tenant_id` = ? and `symbol_id` = ? and `margin_mode` = ?", tenantId, symbolId, marginMode)
		return err
	})
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(oldRows)+len(oldDefaultIds)+2)
	for _, row := range oldRows {
		keys = append(keys, fmt.Sprintf("%s%v", cacheTTradeSymbolLeverageConfigIdPrefix, row.Id))
	}
	keys = append(keys, fmt.Sprintf("%s%v:%v:%v", cacheTTradeSymbolLeverageConfigTenantIdSymbolIdMarginModePrefix, tenantId, symbolId, marginMode))
	for _, id := range oldDefaultIds {
		keys = append(keys, fmt.Sprintf("%s%v", cacheTTradeSymbolLeverageDefaultIdPrefix, id))
	}
	keys = append(keys, fmt.Sprintf("%s%v:%v:%v", cacheTTradeSymbolLeverageDefaultTenantIdSymbolIdMarginModePrefix, tenantId, symbolId, marginMode))
	return m.DelCacheCtx(ctx, keys...)
}

func NewTTradeSymbolLeverageConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolLeverageConfigModel {
	return &customTTradeSymbolLeverageConfigModel{
		defaultTTradeSymbolLeverageConfigModel: newTTradeSymbolLeverageConfigModel(conn, c, opts...),
	}
}

func (m *defaultTTradeSymbolLeverageConfigModel) FindPage(ctx context.Context, filter TradeSymbolLeverageConfigPageFilter, cursor int64, limit int64) ([]*TTradeSymbolLeverageConfig, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("tenant_id", filter.TenantId)
	builder.EqInt64("symbol_id", filter.SymbolId)
	builder.EqInt64("margin_mode", filter.MarginMode)
	if filter.Enabled > 0 {
		builder.And("enabled = ?", filter.Enabled)
	}

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
			`SELECT %s FROM %s WHERE %s ORDER BY sort ASC, id DESC LIMIT ?`,
			tTradeSymbolLeverageConfigRows, m.table, where,
		)
		listArgs = append(listArgs, limit)
	} else {
		listSql = fmt.Sprintf(
			`SELECT %s FROM %s WHERE %s AND id < ? ORDER BY sort ASC, id DESC LIMIT ?`,
			tTradeSymbolLeverageConfigRows, m.table, where,
		)
		listArgs = append(listArgs, cursor, limit)
	}

	var list []*TTradeSymbolLeverageConfig
	if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
