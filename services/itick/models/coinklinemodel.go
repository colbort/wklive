package models

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var _ CoinKlineModel = (*customCoinKlineModel)(nil)

type (
	// CoinKlineModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCoinKlineModel.
	CoinKlineModel interface {
		coinKlineModel
	}

	customCoinKlineModel struct {
		*defaultCoinKlineModel
	}
)

// NewCoinKlineModel returns a model for the mongo.
func NewCoinKlineModel(url, db, collection string) CoinKlineModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customCoinKlineModel{
		defaultCoinKlineModel: newDefaultCoinKlineModel(conn),
	}
}

func KlineCollectionName(market, interval string) string {
	market = normalizeMarket(market)
	interval = normalizeInterval(interval)
	if market == "" || interval == "" {
		return ""
	}
	return fmt.Sprintf("%s_kline_%s", market, interval)
}

type CoinKlineModelFactory struct {
	uri      string
	database string
}

func NewCoinKlineModelFactory(uri, database string) *CoinKlineModelFactory {
	return &CoinKlineModelFactory{
		uri:      uri,
		database: database,
	}
}

func (f *CoinKlineModelFactory) New(market, interval string) CoinKlineModel {
	collection := KlineCollectionName(market, interval)
	return NewCoinKlineModel(f.uri, f.database, collection)
}

func (m *defaultCoinKlineModel) UpsertBySymbolTs(ctx context.Context, data *CoinKline) error {
	if data == nil {
		return fmt.Errorf("coin kline is nil")
	}

	data.Normalize()

	now := time.Now()
	if data.CreateAt.IsZero() {
		data.CreateAt = now
	}
	data.UpdateAt = now

	filter := bson.M{
		"symbol": data.Symbol,
		"ts":     data.Ts,
	}

	update := bson.M{
		"$set": bson.M{
			"market":   data.Market,
			"symbol":   data.Symbol,
			"interval": data.Interval,
			"ts":       data.Ts,
			"openTime": data.OpenTime,
			"open":     data.Open,
			"high":     data.High,
			"low":      data.Low,
			"close":    data.Close,
			"volume":   data.Volume,
			"turnover": data.Turnover,
			"updateAt": data.UpdateAt,
		},
		"$setOnInsert": bson.M{
			"createAt": data.CreateAt,
		},
	}

	_, err := m.conn.UpdateOne(ctx, filter, update)
	return err
}

func (m *defaultCoinKlineModel) FindLatestBySymbol(ctx context.Context, symbol string, limit int64) ([]*CoinKline, error) {
	if limit <= 0 {
		limit = 100
	}

	filter := bson.M{
		"symbol": normalizeSymbol(symbol),
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "ts", Value: -1}}).
		SetLimit(limit)

	var list []*CoinKline
	err := m.conn.Find(ctx, &list, filter, opts)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (m *defaultCoinKlineModel) FindBeforeTsBySymbol(ctx context.Context, symbol string, beforeTs int64, limit int64) ([]*CoinKline, error) {
	if limit <= 0 {
		limit = 100
	}

	filter := bson.M{
		"symbol": normalizeSymbol(symbol),
		"ts": bson.M{
			"$lt": beforeTs,
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "ts", Value: -1}}).
		SetLimit(limit)

	var list []*CoinKline
	err := m.conn.Find(ctx, &list, filter, opts)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (m *defaultCoinKlineModel) EnsureIndexes(ctx context.Context) ([]string, error) {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "symbol", Value: 1},
				{Key: "ts", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("uk_symbol_ts"),
		},
		{
			Keys: bson.D{
				{Key: "symbol", Value: 1},
				{Key: "ts", Value: -1},
			},
			Options: options.Index().SetName("idx_symbol_ts_desc"),
		},
	}

	m.conn.Collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "symbol", Value: 1},
			{Key: "ts", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("uk_symbol_ts"),
	})
	return m.conn.Collection.Indexes().CreateMany(ctx, indexes)
}
