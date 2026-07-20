package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeSymbolSecondsModel = (*customTTradeSymbolSecondsModel)(nil)

type (
	// TTradeSymbolSecondsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeSymbolSecondsModel.
	TTradeSymbolSecondsModel interface {
		tTradeSymbolSecondsModel
		FindAllByTenantIdSymbolId(ctx context.Context, tenantId, symbolId int64) ([]*TTradeSymbolSeconds, error)
	}

	customTTradeSymbolSecondsModel struct {
		*defaultTTradeSymbolSecondsModel
	}
)

func (m *customTTradeSymbolSecondsModel) FindAllByTenantIdSymbolId(ctx context.Context, tenantId, symbolId int64) ([]*TTradeSymbolSeconds, error) {
	query := fmt.Sprintf("select %s from %s where `tenant_id` = ? and `symbol_id` = ? order by `duration_seconds` asc", tTradeSymbolSecondsRows, m.table)
	var rows []*TTradeSymbolSeconds
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, tenantId, symbolId); err != nil {
		return nil, err
	}
	return rows, nil
}

// NewTTradeSymbolSecondsModel returns a model for the database table.
func NewTTradeSymbolSecondsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeSymbolSecondsModel {
	return &customTTradeSymbolSecondsModel{
		defaultTTradeSymbolSecondsModel: newTTradeSymbolSecondsModel(conn, c, opts...),
	}
}
