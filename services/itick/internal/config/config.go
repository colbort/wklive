package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Itick ItickConf
}

type ItickConf struct {
	ApiUrl string
	WSUrl  string
	Token  string
}
