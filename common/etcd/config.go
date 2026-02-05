package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	configurator "github.com/zeromicro/go-zero/core/configcenter"
	"github.com/zeromicro/go-zero/core/configcenter/subscriber"
	v3 "go.etcd.io/etcd/client/v3"
)

func LoadFromEtcdAndMerge(hosts []string, key string, c any) {
	// 1) 这里的 etcd 地址你也可以写到本地 yaml 里，例如 c.EtcdConf.Hosts
	cli, err := v3.New(v3.Config{
		Endpoints:   hosts,
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		// etcd 不可用就降级用本地配置
		return
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, key)
	if err != nil || len(resp.Kvs) == 0 {
		return
	}

	// 2) value 里放的是 yaml（推荐），直接用 conf.LoadFromYamlBytes 覆盖到 c
	data := resp.Kvs[0].Value
	// 注意：go-zero core/conf 有 LoadFromYamlBytes（不同版本函数名可能略有差异）
	_ = conf.LoadFromYamlBytes(data, c) // 覆盖/合并到结构体
}

func WatcherConfig[T any](hosts []string, key string) {
	go func() {
		ss := subscriber.MustNewEtcdSubscriber(subscriber.EtcdConf{
			Hosts: hosts, // etcd 地址
			Key:   key,   // 配置key
		})

		// 创建 configurator
		cc := configurator.MustNewConfigCenter[T](configurator.Config{
			Type: "yaml", // 配置值类型：json,yaml,toml
		}, ss)

		// 获取配置
		// 注意: 配置如果发生变更，调用的结果永远获取到最新的配置
		// v, err := cc.GetConfig()
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println(v)

		// 如果想监听配置变化，可以添加 listener
		cc.AddListener(func() {
			v, err := cc.GetConfig()
			if err != nil {
				panic(err)
			}
			fmt.Println(v)
		})

		select {}
	}()
}
