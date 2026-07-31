package models

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMarginLotApplicationModel = (*customTOptionMarginLotApplicationModel)(nil)

type (
	// TOptionMarginLotApplicationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMarginLotApplicationModel.
	TOptionMarginLotApplicationModel interface {
		tOptionMarginLotApplicationModel
		InsertIgnore(ctx context.Context, item *TOptionMarginLotApplication) (bool, error)
	}

	customTOptionMarginLotApplicationModel struct {
		*defaultTOptionMarginLotApplicationModel
	}
)

func (m *defaultTOptionMarginLotApplicationModel) InsertIgnore(
	ctx context.Context, item *TOptionMarginLotApplication,
) (bool, error) {
	const query = `INSERT IGNORE INTO t_option_margin_lot_application
(tenant_id,instruction_id,margin_lot_id,action,amount,create_times)
VALUES (?,?,?,?,?,?)`
	result, err := m.ExecNoCacheCtx(ctx, query,
		item.TenantId, item.InstructionId, item.MarginLotId, item.Action,
		item.Amount, item.CreateTimes,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

// NewTOptionMarginLotApplicationModel returns a model for the database table.
func NewTOptionMarginLotApplicationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMarginLotApplicationModel {
	return &customTOptionMarginLotApplicationModel{
		defaultTOptionMarginLotApplicationModel: newTOptionMarginLotApplicationModel(conn, c, opts...),
	}
}
