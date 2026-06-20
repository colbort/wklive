package svc

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"wklive/proto/system"
	"wklive/services/user/internal/config"
	"wklive/services/user/models"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config            config.Config
	DB                sqlx.SqlConn
	Redis             *redis.Redis
	Node              *snowflake.Node
	UserModel         models.TUserModel
	UserSecurityModel models.TUserSecurityModel
	UserIdentityModel models.TUserIdentityModel
	UserBankModel     models.TUserBankModel
	FingerprintModel  models.TUserFingerprintModel
	SystemCli         system.SystemClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	node, err := snowflake.NewNode(int64(c.DevServer.Port))
	if err != nil {
		panic(err)
	}
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	systemCli := zrpc.MustNewClient(c.SystemRpc)
	return &ServiceContext{
		Config:            c,
		DB:                conn,
		Redis:             redis.MustNewRedis(c.Redis.RedisConf),
		Node:              node,
		UserModel:         models.NewTUserModel(conn, c.CacheRedis),
		UserSecurityModel: models.NewTUserSecurityModel(conn, c.CacheRedis),
		UserIdentityModel: models.NewTUserIdentityModel(conn, c.CacheRedis),
		UserBankModel:     models.NewTUserBankModel(conn, c.CacheRedis),
		FingerprintModel:  models.NewTUserFingerprintModel(conn, c.CacheRedis),
		SystemCli:         system.NewSystemClient(systemCli.Conn()),
	}
}

func (s *ServiceContext) GenerateInviteCode(ctx context.Context, tenantId int64) (string, error) {
	const (
		codeLength = 6
		maxRetries = 12
		alphabet   = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	)

	for i := 0; i < maxRetries; i++ {
		code, err := randomInviteCode(codeLength, alphabet)
		if err != nil {
			return "", err
		}

		_, err = s.UserModel.FindOneByTenantIdInviteCode(ctx, tenantId, sql.NullString{
			String: code,
			Valid:  true,
		})
		if errors.Is(err, models.ErrNotFound) {
			return code, nil
		}
		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("failed to generate unique invite code")
}

func randomInviteCode(length int, alphabet string) (string, error) {
	code := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))
	for i := range code {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = alphabet[n.Int64()]
	}
	return string(code), nil
}
