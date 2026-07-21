package config

import (
	"wklive/common/mq/kafka"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	CacheRedis cache.CacheConf
	Mysql      struct {
		DataSource string
	} `json:"Mysql" yaml:"Mysql"`
	Jwt struct {
		AccessSecret string
		AccessExpire int64
	} `json:"Jwt" yaml:"Jwt"`
	Register struct {
		UsernameNoRechargeLimit int64
	} `json:"Register" yaml:"Register"`
	GuestTransfer struct {
		ExpireSeconds int64
	} `json:"GuestTransfer" yaml:"GuestTransfer"`
	SystemRpc zrpc.RpcClientConf
	MQ        mq.Config
}
