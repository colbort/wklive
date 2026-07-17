package main

import (
	"net/http"

	"wklive/admin-api/internal/config"
	"wklive/common/reqenc"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func mustNewRequestEncryption(c config.Config) *reqenc.Service {
	encryptionConfig := c.RequestEncryption.WithDefaults()
	if encryptionConfig.Scope == "" {
		encryptionConfig.Scope = "admin-api"
	}
	var store reqenc.Store
	if encryptionConfig.Mode != reqenc.ModeDisabled {
		store = reqenc.NewRedisStore(redis.MustNewRedis(c.RedisConf), encryptionConfig.Scope)
	}
	service, err := reqenc.New(encryptionConfig, store)
	if err != nil {
		panic(err)
	}
	return service
}

func adminRequestEncryptionRegistry() *reqenc.Registry {
	prefixes := []string{
		"/admin/asset/",
		"/admin/itick/",
		"/admin/member/",
		"/admin/option/",
		"/admin/payment/",
		"/admin/staking/",
		"/admin/system/",
		"/admin/trade/",
	}
	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch}
	rules := make([]reqenc.Rule, 0, len(prefixes)*len(methods))
	for _, prefix := range prefixes {
		for _, method := range methods {
			rules = append(rules, reqenc.Rule{
				Method: method, Path: prefix, PathPrefix: true, Location: reqenc.LocationJSON,
			})
		}
	}
	return reqenc.NewRegistry(rules...)
}
