package svc

import (
	"wklive/services/chat/internal/config"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                config.Config
	DB                    sqlx.SqlConn
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

	return &ServiceContext{
		Config:                c,
		DB:                    conn,
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
}
