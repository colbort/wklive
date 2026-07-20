package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TTradeAssetReservationModel = (*customTTradeAssetReservationModel)(nil)

type (
	// TTradeAssetReservationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTTradeAssetReservationModel.
	TTradeAssetReservationModel interface {
		tTradeAssetReservationModel
	}

	customTTradeAssetReservationModel struct {
		*defaultTTradeAssetReservationModel
	}
)

// NewTTradeAssetReservationModel returns a model for the database table.
func NewTTradeAssetReservationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TTradeAssetReservationModel {
	return &customTTradeAssetReservationModel{
		defaultTTradeAssetReservationModel: newTTradeAssetReservationModel(conn, c, opts...),
	}
}
