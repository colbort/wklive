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
		FindPortfolioActiveByAccount(ctx context.Context, tenantId, userId, accountId int64, settleCoin string) ([]*TOptionMarginLot, error)
		FindOneForUpdate(ctx context.Context, id int64) (*TOptionMarginLot, error)
	}

	customTOptionMarginLotModel struct {
		*defaultTOptionMarginLotModel
	}
)

func (m *defaultTOptionMarginLotModel) FindPortfolioActiveByAccount(
	ctx context.Context,
	tenantId, userId, accountId int64,
	settleCoin string,
) ([]*TOptionMarginLot, error) {
	accountClause := ""
	args := []any{tenantId, userId}
	if accountId > 0 {
		accountClause = " AND l.account_id = ?"
		args = append(args, accountId)
	}
	query := fmt.Sprintf(`SELECT l.* FROM %s l
JOIN t_option_contract c ON c.id = l.contract_id AND c.tenant_id = l.tenant_id
WHERE l.tenant_id = ? AND l.user_id = ?%s
  AND c.settle_coin = ? AND c.seller_margin_mode = ?
  AND l.status IN (?, ?) AND l.remaining_margin > l.pending_margin
ORDER BY l.id FOR UPDATE`, m.table, accountClause)
	var list []*TOptionMarginLot
	args = append(args,
		settleCoin,
		int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO),
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING),
	)
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...)
	return list, err
}

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
WHERE tenant_id = ? AND position_id = ? AND status IN (?, ?, ?)
  AND remaining_quantity > 0 AND remaining_margin > pending_margin
ORDER BY id FOR UPDATE`, tOptionMarginLotRows, m.table)
	var list []*TOptionMarginLot
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, tenantId, positionId,
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING),
		int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING),
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
