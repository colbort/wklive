package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionUserTradingControlModel = (*customTOptionUserTradingControlModel)(nil)

type (
	// TOptionUserTradingControlModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionUserTradingControlModel.
	TOptionUserTradingControlModel interface {
		tOptionUserTradingControlModel
		EnsureForUpdate(ctx context.Context, tenantId, userId, now int64) (*TOptionUserTradingControl, error)
		FindForUpdate(ctx context.Context, tenantId, userId int64) (*TOptionUserTradingControl, error)
	}

	customTOptionUserTradingControlModel struct {
		*defaultTOptionUserTradingControlModel
	}
)

func (m *customTOptionUserTradingControlModel) FindForUpdate(
	ctx context.Context, tenantId, userId int64,
) (*TOptionUserTradingControl, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND user_id = ? LIMIT 1 FOR UPDATE`, tOptionUserTradingControlRows, m.table)
	var item TOptionUserTradingControl
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, tenantId, userId); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *customTOptionUserTradingControlModel) EnsureForUpdate(
	ctx context.Context, tenantId, userId, now int64,
) (*TOptionUserTradingControl, error) {
	if _, err := m.ExecNoCacheCtx(ctx, `INSERT IGNORE INTO t_option_user_trading_control
(tenant_id,user_id,kill_switch,create_times,update_times) VALUES (?,?,2,?,?)`,
		tenantId, userId, now, now); err != nil {
		return nil, err
	}
	return m.FindForUpdate(ctx, tenantId, userId)
}

// NewTOptionUserTradingControlModel returns a model for the database table.
func NewTOptionUserTradingControlModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionUserTradingControlModel {
	return &customTOptionUserTradingControlModel{
		defaultTOptionUserTradingControlModel: newTOptionUserTradingControlModel(conn, c, opts...),
	}
}
