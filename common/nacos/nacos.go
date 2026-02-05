package nacos

import (
	nacosconf "wklive/common/nacos/config"
	"wklive/common/nacos/naming"
	"wklive/common/nacos/types"
)

func NewRegistry(c types.NacosConf) (*naming.Registry, error)   { return naming.NewRegistry(c) }
func NewDiscovery(c types.NacosConf) (*naming.Discovery, error) { return naming.NewDiscovery(c) }

func RegisterGrpcResolver(c types.NacosConf) error { return naming.RegisterResolver(c) }

func NewConfigSubscriber(c types.NacosConf) (*nacosconf.Subscriber, error) {
	return nacosconf.NewSubscriber(c)
}

func NewConfigLoader(c types.NacosConf) (*nacosconf.Loader, error) {
	return nacosconf.NewLoader(c)
}

