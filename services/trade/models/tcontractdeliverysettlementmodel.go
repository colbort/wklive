package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractDeliverySettlementModel = (*customTContractDeliverySettlementModel)(nil)

type (
	// TContractDeliverySettlementModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractDeliverySettlementModel.
	TContractDeliverySettlementModel interface {
		tContractDeliverySettlementModel
	}

	customTContractDeliverySettlementModel struct {
		*defaultTContractDeliverySettlementModel
	}
)

// NewTContractDeliverySettlementModel returns a model for the database table.
func NewTContractDeliverySettlementModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractDeliverySettlementModel {
	return &customTContractDeliverySettlementModel{
		defaultTContractDeliverySettlementModel: newTContractDeliverySettlementModel(conn, c, opts...),
	}
}
