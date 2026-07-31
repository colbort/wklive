package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionCorporateActionMarginLotModel = (*customTOptionCorporateActionMarginLotModel)(nil)

type (
	// TOptionCorporateActionMarginLotModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionCorporateActionMarginLotModel.
	TOptionCorporateActionMarginLotModel interface {
		tOptionCorporateActionMarginLotModel
	}

	customTOptionCorporateActionMarginLotModel struct {
		*defaultTOptionCorporateActionMarginLotModel
	}
)

// NewTOptionCorporateActionMarginLotModel returns a model for the database table.
func NewTOptionCorporateActionMarginLotModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionCorporateActionMarginLotModel {
	return &customTOptionCorporateActionMarginLotModel{
		defaultTOptionCorporateActionMarginLotModel: newTOptionCorporateActionMarginLotModel(conn, c, opts...),
	}
}
