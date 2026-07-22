package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolSessionModel = (*customTTradeSymbolSessionModel)(nil)

type (
	// TTradeSymbolSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolSessionModel.
	TTradeSymbolSessionModel interface {
		tTradeSymbolSessionModel
		FindAllByTenantIdSymbolId(ctx context.Context, tenantId, symbolId int64) ([]*TTradeSymbolSession, error)
	}

	customTTradeSymbolSessionModel struct {
		*defaultTTradeSymbolSessionModel
	}
)

func (m *customTTradeSymbolSessionModel) FindAllByTenantIdSymbolId(ctx context.Context, tenantId, symbolId int64) ([]*TTradeSymbolSession, error) {
	query := fmt.Sprintf("select %s from %s where `tenant_id` = ? and `symbol_id` = ? order by `day_of_week` asc, `start_second` asc", tTradeSymbolSessionRows, m.table)
	var rows []*TTradeSymbolSession
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, tenantId, symbolId); err != nil {
		return nil, err
	}
	return rows, nil
}

// NewTTradeSymbolSessionModel returns a model for the database table.
func NewTTradeSymbolSessionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolSessionModel {
	return &customTTradeSymbolSessionModel{
		defaultTTradeSymbolSessionModel: newTTradeSymbolSessionModel(conn, c, opts...),
	}
}
