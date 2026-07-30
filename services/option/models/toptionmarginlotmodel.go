package models

import (
	"context"
	"fmt"

	"wklive/proto/option"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionMarginLotModel = (*customTOptionMarginLotModel)(nil)

type (
	// TOptionMarginLotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionMarginLotModel.
	TOptionMarginLotModel interface {
		tOptionMarginLotModel
		FindActiveByPosition(ctx context.Context, tenantId, positionId int64) ([]*TOptionMarginLot, error)
		FindClosableByPosition(ctx context.Context, tenantId, positionId int64) ([]*TOptionMarginLot, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionMarginLot, error)
	}

	customTOptionMarginLotModel struct {
		*defaultTOptionMarginLotModel
	}
)

func (m *defaultTOptionMarginLotModel) FindOneForUpdate(ctx context.Context, id int64) (*TOptionMarginLot, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? LIMIT 1 FOR UPDATE", tOptionMarginLotRows, m.table)
	var item TOptionMarginLot
	if err := m.QueryRowNoCacheCtx(ctx, &item, query, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *defaultTOptionMarginLotModel) FindClosableByPosition(ctx context.Context, tenantId, positionId int64) ([]*TOptionMarginLot, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND position_id = ? AND status IN (?, ?)
  AND remaining_quantity > 0 AND remaining_margin > pending_margin
ORDER BY id FOR UPDATE`, tOptionMarginLotRows, m.table)
	var list []*TOptionMarginLot
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, positionId,
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING),
	)
	return list, err
}

func (m *defaultTOptionMarginLotModel) FindActiveByPosition(ctx context.Context, tenantId, positionId int64) ([]*TOptionMarginLot, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND position_id = ? AND status IN (?, ?) AND remaining_margin > pending_margin
ORDER BY id FOR UPDATE`, tOptionMarginLotRows, m.table)
	var list []*TOptionMarginLot
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, positionId,
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING),
	)
	return list, err
}

// NewTOptionMarginLotModel returns a model for the database table.
func NewTOptionMarginLotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionMarginLotModel {
	return &customTOptionMarginLotModel{
		defaultTOptionMarginLotModel: newTOptionMarginLotModel(conn, c, opts...),
	}
}
