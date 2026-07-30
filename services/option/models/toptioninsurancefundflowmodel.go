package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TOptionInsuranceFundFlowModel = (*customTOptionInsuranceFundFlowModel)(nil)

type (
	// TOptionInsuranceFundFlowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTOptionInsuranceFundFlowModel.
	TOptionInsuranceFundFlowModel interface {
		tOptionInsuranceFundFlowModel
	}

	customTOptionInsuranceFundFlowModel struct {
		*defaultTOptionInsuranceFundFlowModel
	}
)

// NewTOptionInsuranceFundFlowModel returns a model for the database table.
func NewTOptionInsuranceFundFlowModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TOptionInsuranceFundFlowModel {
	return &customTOptionInsuranceFundFlowModel{
		defaultTOptionInsuranceFundFlowModel: newTOptionInsuranceFundFlowModel(conn, c, opts...),
	}
}
