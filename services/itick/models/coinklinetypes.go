package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CoinKline struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// 市场，例如：binance / okx
	Market string `bson:"market,omitempty" json:"market,omitempty"`

	// 交易对，例如：BTCUSDT
	Symbol string `bson:"symbol,omitempty" json:"symbol,omitempty"`

	// 周期，例如：1m / 5m / 15m / 1h / 1d
	Interval string `bson:"interval,omitempty" json:"interval,omitempty"`

	// K线开始时间戳，毫秒
	Ts int64 `bson:"ts,omitempty" json:"ts,omitempty"`

	// 开盘时间
	OpenTime time.Time `bson:"openTime,omitempty" json:"openTime,omitempty"`

	Open  float64 `bson:"open,omitempty" json:"open,omitempty"`
	High  float64 `bson:"high,omitempty" json:"high,omitempty"`
	Low   float64 `bson:"low,omitempty" json:"low,omitempty"`
	Close float64 `bson:"close,omitempty" json:"close,omitempty"`

	// 成交量
	Volume float64 `bson:"volume,omitempty" json:"volume,omitempty"`

	// 成交额
	Turnover float64 `bson:"turnover,omitempty" json:"turnover,omitempty"`

	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}

func (m *CoinKline) Normalize() {
	m.Market = normalizeMarket(m.Market)
	m.Symbol = normalizeSymbol(m.Symbol)
	m.Interval = normalizeInterval(m.Interval)
}

func normalizeMarket(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeSymbol(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

func normalizeInterval(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
