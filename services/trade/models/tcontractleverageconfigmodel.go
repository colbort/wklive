package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractLeverageConfigModel = (*customTContractLeverageConfigModel)(nil)

type (
	// TContractLeverageConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractLeverageConfigModel.
	TContractLeverageConfigModel interface {
		tContractLeverageConfigModel
	}

	customTContractLeverageConfigModel struct {
		*defaultTContractLeverageConfigModel
	}
)

// NewTContractLeverageConfigModel returns a model for the database table.
func NewTContractLeverageConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractLeverageConfigModel {
	return &customTContractLeverageConfigModel{
		defaultTContractLeverageConfigModel: newTContractLeverageConfigModel(conn, c, opts...),
	}
}
