package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionTradeCorrectionLegModel = (*customTOptionTradeCorrectionLegModel)(nil)

type (
	// TOptionTradeCorrectionLegModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionTradeCorrectionLegModel.
	TOptionTradeCorrectionLegModel interface {
		tOptionTradeCorrectionLegModel
		FindByCorrection(ctx context.Context, tenantId, correctionId int64) ([]*TOptionTradeCorrectionLeg, error)
	}

	customTOptionTradeCorrectionLegModel struct {
		*defaultTOptionTradeCorrectionLegModel
	}
)

func (m *defaultTOptionTradeCorrectionLegModel) FindByCorrection(
	ctx context.Context, tenantId, correctionId int64,
) ([]*TOptionTradeCorrectionLeg, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id = ? AND correction_id = ? ORDER BY leg_no ASC`, tOptionTradeCorrectionLegRows, m.table)
	var items []*TOptionTradeCorrectionLeg
	if err := m.QueryRowsNoCacheCtx(ctx, &items, query, tenantId, correctionId); err != nil {
		return nil, err
	}
	return items, nil
}

// NewTOptionTradeCorrectionLegModel returns a model for the database table.
func NewTOptionTradeCorrectionLegModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionTradeCorrectionLegModel {
	return &customTOptionTradeCorrectionLegModel{
		defaultTOptionTradeCorrectionLegModel: newTOptionTradeCorrectionLegModel(conn, c, opts...),
	}
}
