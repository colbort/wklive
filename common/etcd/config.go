package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	configurator "github.com/zeromicro/go-zero/core/configcenter"
	"github.com/zeromicro/go-zero/core/configcenter/subscriber"
	"github.com/zeromicro/go-zero/core/logx"
	v3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v2"
)

func LoadFromEtcdAndMerge(hosts []string, keys []string, c any) {
	cli, err := v3.New(v3.Config{
		Endpoints:   hosts,
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		return
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	merged := make(map[string]any)

	for _, key := range keys {
		resp, err := cli.Get(ctx, key)
		if err != nil || len(resp.Kvs) == 0 {
			continue
		}

		data := resp.Kvs[0].Value

		var m map[string]any
		if err := yaml.Unmarshal(data, &m); err != nil {
			logx.Errorf("yaml parse failed key=%s err=%v", key, err)
			continue
		}

		deepMerge(merged, m)
	}

	// 最后一次性 decode 到 struct
	bs, _ := yaml.Marshal(merged)
	if err := conf.LoadFromYamlBytes(bs, c); err != nil {
		logx.Errorf("load merged yaml failed err=%v yaml=%s", err, string(bs))
	}
}

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if dstMap, ok2 := dst[k].(map[string]any); ok2 {
				deepMerge(dstMap, vMap)
			} else {
				dst[k] = vMap
			}
		} else {
			dst[k] = v
		}
	}
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
