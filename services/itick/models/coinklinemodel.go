package models

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var _ CoinKlineModel = (*customCoinKlineModel)(nil)

type (
	// 这里假设 coinKlineModel / defaultCoinKlineModel / CoinKline
	// 是你 goctl 生成的原有内容
	CoinKlineModel interface {
		coinKlineModel

		UpsertBySymbolTs(ctx context.Context, data *CoinKline) error
		BulkUpsertBySymbolTs(ctx context.Context, list []*CoinKline) error

		FindLatestBySymbol(ctx context.Context, symbol string, limit int64) ([]*CoinKline, error)
		FindBeforeTsBySymbol(ctx context.Context, symbol string, beforeTs int64, limit int64) ([]*CoinKline, error)
		EnsureIndexes(ctx context.Context) ([]string, error)
	}

	customCoinKlineModel struct {
		*defaultCoinKlineModel
	}
)

// NewCoinKlineModel returns a model for mongo.
func NewCoinKlineModel(url, db, collection string) CoinKlineModel {
	if collection == "" {
		panic("mongo collection is empty")
	}

	conn := mon.MustNewModel(url, db, collection)
	return &customCoinKlineModel{
		defaultCoinKlineModel: newDefaultCoinKlineModel(conn),
	}
}

func KlineCollectionName(categoryCode, interval string) string {
	categoryCode = normalizeCategory(categoryCode)
	interval = normalizeInterval(interval)
	if categoryCode == "" || interval == "" {
		return ""
	}
	return fmt.Sprintf("%s_kline_%s", categoryCode, interval)
}

type CoinKlineModelFactory struct {
	uri      string
	database string

	mu     sync.RWMutex
	models map[string]CoinKlineModel
}

func NewCoinKlineModelFactory(uri, database string) *CoinKlineModelFactory {
	return &CoinKlineModelFactory{
		uri:      uri,
		database: database,
		models:   make(map[string]CoinKlineModel),
	}
}

func (f *CoinKlineModelFactory) cacheKey(categoryCode, interval string) string {
	return fmt.Sprintf("%s:%s", normalizeCategory(categoryCode), normalizeInterval(interval))
}

// New 保留原来的调用方式，但内部做缓存
func (f *CoinKlineModelFactory) New(categoryCode, interval string) CoinKlineModel {
	key := f.cacheKey(categoryCode, interval)
	collection := KlineCollectionName(categoryCode, interval)
	if key == ":" || collection == "" {
		return nil
	}

	f.mu.RLock()
	model, ok := f.models[key]
	f.mu.RUnlock()
	if ok {
		return model
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if model, ok := f.models[key]; ok {
		return model
	}

	model = NewCoinKlineModel(f.uri, f.database, collection)
	f.models[key] = model
	return model
}

func (f *CoinKlineModelFactory) WarmupAndEnsureIndexes(ctx context.Context, categoryCode, interval string) error {
	model := f.New(categoryCode, interval)
	if model == nil {
		return fmt.Errorf("invalid categoryCode or interval, categoryCode=%q interval=%q", categoryCode, interval)
	}

	_, err := model.EnsureIndexes(ctx)
	return err
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
			"categoryCode": data.CategoryCode,
			"market":       data.Market,
			"symbol":       data.Symbol,
			"interval":     data.Interval,
			"ts":           data.Ts,
			"open":         data.Open,
			"high":         data.High,
			"low":          data.Low,
			"close":        data.Close,
			"volume":       data.Volume,
			"turnover":     data.Turnover,
			"updateAt":     data.UpdateAt,
		},
		"$setOnInsert": bson.M{
			"createAt": data.CreateAt,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := m.conn.UpdateOne(ctx, filter, update, opts)
	return err
}

func (m *defaultCoinKlineModel) BulkUpsertBySymbolTs(ctx context.Context, list []*CoinKline) error {
	if len(list) == 0 {
		return nil
	}

	now := time.Now()
	writes := make([]mongo.WriteModel, 0, len(list))

	for _, data := range list {
		if data == nil {
			continue
		}

		data.Normalize()
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
				"categoryCode": data.CategoryCode,
				"market":       data.Market,
				"symbol":       data.Symbol,
				"interval":     data.Interval,
				"ts":           data.Ts,
				"open":         data.Open,
				"high":         data.High,
				"low":          data.Low,
				"close":        data.Close,
				"volume":       data.Volume,
				"turnover":     data.Turnover,
				"updateAt":     data.UpdateAt,
			},
			"$setOnInsert": bson.M{
				"createAt": data.CreateAt,
			},
		}

		writes = append(writes, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}

	if len(writes) == 0 {
		return nil
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := m.conn.Collection.BulkWrite(ctx, writes, opts)
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
	return m.conn.Collection.Indexes().CreateMany(ctx, indexes)
}
