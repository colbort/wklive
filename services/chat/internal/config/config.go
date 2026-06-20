package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Mysql struct {
		DataSource string
	} `json:"Mysql" yaml:"Mysql"`
	Mongo struct {
		Url string
		Db  string
	}
}
