package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractFundingSettlementModel = (*customTContractFundingSettlementModel)(nil)

type (
	// TContractFundingSettlementModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractFundingSettlementModel.
	TContractFundingSettlementModel interface {
		tContractFundingSettlementModel
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
