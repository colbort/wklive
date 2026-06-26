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

var _ ChatMessageModel = (*defaultChatMessageModel)(nil)

type (
	ChatMessagePageFilter struct {
		MerchantId  int64
		SessionNo   string
		UserId      int64
		AgentId     int64
		SenderType  int64
		MessageType int64
		BeforeTime  int64
		AfterTime   int64
	}

	ChatMessageModel interface {
		Insert(ctx context.Context, data *ChatMessage) error
		UpsertByMessageNo(ctx context.Context, data *ChatMessage) error
		FindOneByMessageNo(ctx context.Context, messageNo string) (*ChatMessage, error)
		FindPage(ctx context.Context, filter ChatMessagePageFilter, limit int64) ([]*ChatMessage, error)
		UpdateStatus(ctx context.Context, merchantId int64, messageNo string, status int64, updateTimes int64) error
		MarkRead(ctx context.Context, merchantId int64, messageNo string, readTime int64, updateTimes int64) error
		EnsureIndexes(ctx context.Context) ([]string, error)
	}

	defaultChatMessageModel struct {
		conn *mon.Model
	}

	ChatMessageModelFactory struct {
		url string
		db  string

		mu     sync.RWMutex
		models map[int64]ChatMessageModel
	}
)

func NewChatMessageModel(url, db, collection string) ChatMessageModel {
	if collection == "" {
		panic("mongo collection is empty")
	}

	conn := mon.MustNewModel(url, db, collection)
	return &defaultChatMessageModel{conn: conn}
}

func NewChatMessageModelFactory(url, db string) *ChatMessageModelFactory {
	return &ChatMessageModelFactory{
		url:    url,
		db:     db,
		models: make(map[int64]ChatMessageModel),
	}
}

func (f *ChatMessageModelFactory) New(merchantId int64) ChatMessageModel {
	collection := ChatMessageCollectionName(merchantId)
	if collection == "" {
		return nil
	}

	f.mu.RLock()
	model, ok := f.models[merchantId]
	f.mu.RUnlock()
	if ok {
		return model
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if model, ok := f.models[merchantId]; ok {
		return model
	}

	model = NewChatMessageModel(f.url, f.db, collection)
	f.models[merchantId] = model
	return model
}

func (f *ChatMessageModelFactory) EnsureIndexes(ctx context.Context, merchantId int64) ([]string, error) {
	model := f.New(merchantId)
	if model == nil {
		return nil, fmt.Errorf("invalid merchant_id: %d", merchantId)
	}
	return model.EnsureIndexes(ctx)
}

func (m *defaultChatMessageModel) Insert(ctx context.Context, data *ChatMessage) error {
	if data == nil {
		return fmt.Errorf("chat message is nil")
	}
	if data.ID.IsZero() {
		data.ID = bson.NewObjectID()
	}
	now := time.Now().UnixMilli()
	if data.CreateTimes == 0 {
		data.CreateTimes = now
	}
	if data.UpdateTimes == 0 {
		data.UpdateTimes = data.CreateTimes
	}

	_, err := m.conn.InsertOne(ctx, data)
	return err
}

func (m *defaultChatMessageModel) UpsertByMessageNo(ctx context.Context, data *ChatMessage) error {
	if data == nil {
		return fmt.Errorf("chat message is nil")
	}
	if data.MessageNo == "" {
		return fmt.Errorf("message_no is empty")
	}

	now := time.Now().UnixMilli()
	if data.CreateTimes == 0 {
		data.CreateTimes = now
	}
	if data.UpdateTimes == 0 {
		data.UpdateTimes = now
	}

	update := bson.M{
		"$set": bson.M{
			"session_no":   data.SessionNo,
			"merchant_id":  data.MerchantId,
			"sender":       data.Sender,
			"receiver":     data.Receiver,
			"message_type": data.MessageType,
			"content":      data.Content,
			"url":          data.Url,
			"file_name":    data.FileName,
			"mime_type":    data.MimeType,
			"file_size":    data.FileSize,
			"width":        data.Width,
			"height":       data.Height,
			"duration":     data.Duration,
			"status":       data.Status,
			"payload":      data.Payload,
			"read_time":    data.ReadTime,
			"update_times": data.UpdateTimes,
		},
		"$setOnInsert": bson.M{
			"message_no":   data.MessageNo,
			"create_times": data.CreateTimes,
		},
	}

	_, err := m.conn.UpdateOne(ctx, bson.M{"message_no": data.MessageNo}, update, options.UpdateOne().SetUpsert(true))
	return err
}

func (m *defaultChatMessageModel) FindOneByMessageNo(ctx context.Context, messageNo string) (*ChatMessage, error) {
	var data ChatMessage
	err := m.conn.FindOne(ctx, &data, bson.M{"message_no": messageNo})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultChatMessageModel) FindPage(ctx context.Context, filter ChatMessagePageFilter, limit int64) ([]*ChatMessage, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := bson.M{}
	clauses := make([]bson.M, 0, 2)
	if filter.MerchantId != 0 {
		query["merchant_id"] = filter.MerchantId
	}
	if filter.SessionNo != "" {
		query["session_no"] = filter.SessionNo
	}
	if filter.UserId != 0 {
		clauses = append(clauses, bson.M{"$or": []bson.M{
			{"sender.type": 1, "sender.id": filter.UserId},
			{"receiver.type": 1, "receiver.id": filter.UserId},
		}})
	}
	if filter.AgentId != 0 {
		clauses = append(clauses, bson.M{"$or": []bson.M{
			{"sender.type": 2, "sender.id": filter.AgentId},
			{"receiver.type": 2, "receiver.id": filter.AgentId},
		}})
	}
	if filter.SenderType != 0 {
		query["sender.type"] = filter.SenderType
	}
	if filter.MessageType != 0 {
		query["message_type"] = filter.MessageType
	}

	timeRange := bson.M{}
	if filter.BeforeTime != 0 {
		timeRange["$lt"] = filter.BeforeTime
	}
	if filter.AfterTime != 0 {
		timeRange["$gt"] = filter.AfterTime
	}
	if len(timeRange) > 0 {
		query["create_times"] = timeRange
	}
	if len(clauses) > 0 {
		query["$and"] = clauses
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "create_times", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(limit)

	var list []*ChatMessage
	if err := m.conn.Find(ctx, &list, query, opts); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *defaultChatMessageModel) UpdateStatus(ctx context.Context, merchantId int64, messageNo string, status int64, updateTimes int64) error {
	if updateTimes == 0 {
		updateTimes = time.Now().UnixMilli()
	}
	filter := bson.M{"message_no": messageNo}
	if merchantId != 0 {
		filter["merchant_id"] = merchantId
	}
	_, err := m.conn.UpdateOne(ctx, filter, bson.M{"$set": bson.M{
		"status":       status,
		"update_times": updateTimes,
	}})
	return err
}

func (m *defaultChatMessageModel) MarkRead(ctx context.Context, merchantId int64, messageNo string, readTime int64, updateTimes int64) error {
	if readTime == 0 {
		readTime = time.Now().UnixMilli()
	}
	if updateTimes == 0 {
		updateTimes = readTime
	}
	filter := bson.M{"message_no": messageNo}
	if merchantId != 0 {
		filter["merchant_id"] = merchantId
	}
	_, err := m.conn.UpdateOne(ctx, filter, bson.M{"$set": bson.M{
		"read_time":    readTime,
		"update_times": updateTimes,
	}})
	return err
}

func (m *defaultChatMessageModel) EnsureIndexes(ctx context.Context) ([]string, error) {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "message_no", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uk_message_no"),
		},
		{
			Keys:    bson.D{{Key: "session_no", Value: 1}, {Key: "create_times", Value: -1}},
			Options: options.Index().SetName("idx_session_time"),
		},
		{
			Keys:    bson.D{{Key: "merchant_id", Value: 1}, {Key: "sender.type", Value: 1}, {Key: "sender.id", Value: 1}, {Key: "create_times", Value: -1}},
			Options: options.Index().SetName("idx_merchant_user_time"),
		},
		{
			Keys:    bson.D{{Key: "merchant_id", Value: 1}, {Key: "receiver.type", Value: 1}, {Key: "receiver.id", Value: 1}, {Key: "create_times", Value: -1}},
			Options: options.Index().SetName("idx_merchant_agent_time"),
		},
	}
	return m.conn.Collection.Indexes().CreateMany(ctx, indexes)
}
