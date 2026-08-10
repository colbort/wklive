package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CoinKline struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// 市场，例如：binance / okx
	CategoryCode string `bson:"category,omitempty" json:"category,omitempty"`

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

	Source         string `bson:"source,omitempty" json:"source,omitempty"`
	SourcePriority int32  `bson:"sourcePriority,omitempty" json:"-"`
	Revision       int64  `bson:"revision,omitempty" json:"revision,omitempty"`
	IsClosed       bool   `bson:"isClosed" json:"isClosed"`
	Confirmed      bool   `bson:"confirmed" json:"confirmed"`
	ActualCount    int32  `bson:"actualCount,omitempty" json:"actualCount,omitempty"`
	ExpectedCount  int32  `bson:"expectedCount,omitempty" json:"expectedCount,omitempty"`

	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}

func (m *CoinKline) Normalize() {
	m.CategoryCode = normalizeCategory(m.CategoryCode)
	m.Market = normalizeMarket(m.Market)
	m.Symbol = normalizeSymbol(m.Symbol)
	m.Interval = normalizeInterval(m.Interval)
	m.Source = strings.ToLower(strings.TrimSpace(m.Source))
	if m.SourcePriority <= 0 {
		m.SourcePriority = KlineSourcePriority(m.Source)
	}
	if m.Revision <= 0 {
		m.Revision = time.Now().UnixMilli()
	}
}

const (
	KlineSourceRealtime = "realtime"
	KlineSourceDerived  = "derived"
	// KlineSourceExchangeRest identifies a direct exchange REST fallback. It is
	// intentionally lower priority than the configured iTick REST source so a
	// later upstream correction can replace it.
	KlineSourceExchangeRest = "exchange-rest"
	KlineSourceRest         = "rest"
)

func KlineSourcePriority(source string) int32 {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case KlineSourceRest:
		return 300
	case KlineSourceExchangeRest:
		return 250
	case KlineSourceDerived:
		return 200
	case KlineSourceRealtime:
		return 100
	default:
		return 1
	}
}

func normalizeCategory(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeMarket(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

func normalizeSymbol(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

func normalizeInterval(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
