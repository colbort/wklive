package svc

import (
	"context"
	"fmt"
	"time"
	"wklive/services/chat/internal/config"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                config.Config
	DB                    sqlx.SqlConn
	ChatMerchantInfoModel models.TChatMerchantInfoModel
	ChatUserModel         models.TChatUserModel
	ChatAgentModel        models.TChatAgentModel
	ChatSessionModel      models.TChatSessionModel
	ChatAssignmentModel   models.TChatAssignmentModel
	ChatQuickReplyModel   models.TChatQuickReplyModel
	ChatCategoryModel     models.TChatCategoryModel
	ChatSatisfactionModel models.TChatSatisfactionModel
	ChatReadCursorModel   models.TChatReadCursorModel
	ChatGroupModel        models.TChatGroupModel
	ChatWorkOrderModel    models.TChatWorkOrderModel
	ChatMessageFactory    *models.ChatMessageModelFactory
	BusRedis              *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	var busRedis *redis.Redis
	if c.Redis.Host != "" {
		rds, err := redis.NewRedis(c.Redis.RedisConf)
		if err != nil {
			logx.Errorf("chat bus redis init failed: %v", err)
		} else {
			busRedis = rds
		}
	}

	svcCtx := &ServiceContext{
		Config:                c,
		DB:                    conn,
		ChatMerchantInfoModel: models.NewTChatMerchantInfoModel(conn, c.CacheRedis),
		ChatUserModel:         models.NewTChatUserModel(conn, c.CacheRedis),
		ChatAgentModel:        models.NewTChatAgentModel(conn, c.CacheRedis),
		ChatSessionModel:      models.NewTChatSessionModel(conn, c.CacheRedis),
		ChatAssignmentModel:   models.NewTChatAssignmentModel(conn, c.CacheRedis),
		ChatQuickReplyModel:   models.NewTChatQuickReplyModel(conn, c.CacheRedis),
		ChatCategoryModel:     models.NewTChatCategoryModel(conn, c.CacheRedis),
		ChatSatisfactionModel: models.NewTChatSatisfactionModel(conn, c.CacheRedis),
		ChatReadCursorModel:   models.NewTChatReadCursorModel(conn, c.CacheRedis),
		ChatGroupModel:        models.NewTChatGroupModel(conn, c.CacheRedis),
		ChatWorkOrderModel:    models.NewTChatWorkOrderModel(conn, c.CacheRedis),
		ChatMessageFactory:    models.NewChatMessageModelFactory(c.Mongo.Url, c.Mongo.Db),
		BusRedis:              busRedis,
	}

	return svcCtx
}

func (s *ServiceContext) GenerateNo(ctx context.Context, prefix string) (string, error) {
	now := time.Now()
	date := now.Format("20060102")

	// 每天、每个前缀单独计数
	key := fmt.Sprintf("chat:%s:%s", prefix, date)

	seq, err := s.BusRedis.IncrCtx(ctx, key)
	if err != nil {
		return "", err
	}

	// 设置过期时间，避免 Redis 一直堆积旧 key
	// 这里只在第一次创建时设置
	if seq == 1 {
		_ = s.BusRedis.ExpireCtx(ctx, key, 36*int(time.Hour.Seconds()))
	}

	orderID := fmt.Sprintf("%s%s%06d", prefix, date, seq)
	return orderID, nil
}
