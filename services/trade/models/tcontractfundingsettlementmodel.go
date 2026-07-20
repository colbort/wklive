package models

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"wklive/common/sqlutil"
)

var _ TContractFundingSettlementModel = (*customTContractFundingSettlementModel)(nil)

type (
	// TContractFundingSettlementModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractFundingSettlementModel.
	TContractFundingSettlementModel interface {
		tContractFundingSettlementModel
		FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractFundingSettlement, int64, error)
	}

	customTContractFundingSettlementModel struct {
		*defaultTContractFundingSettlementModel
	}
)

// NewTContractFundingSettlementModel returns a model for the database table.
func NewTContractFundingSettlementModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractFundingSettlementModel {
	return &customTContractFundingSettlementModel{
		defaultTContractFundingSettlementModel: newTContractFundingSettlementModel(conn, c, opts...),
	}
}

func (m *defaultTContractFundingSettlementModel) FindPage(ctx context.Context, filter AdminPageFilter, cursor, limit int64) ([]*TContractFundingSettlement, int64, error) {
	limit = sqlutil.NormalizeLimit(limit)
	b := adminPageBuilder(filter, "")
	where, args := b.Where(), b.Args()
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	la := append([]any{}, args...)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", tContractFundingSettlementRows, m.table, where)
	if cursor > 0 {
		q += " AND id < ?"
		la = append(la, cursor)
	}
	q += " ORDER BY id DESC LIMIT ?"
	la = append(la, limit)
	var list []*TContractFundingSettlement
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, la...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
