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

func adminRequestEncryptionRegistry(encryptionConfig reqenc.Config) *reqenc.Registry {
	rules := []reqenc.Rule{
		// The client needs these plaintext endpoints before it has an AES session.
		{Method: http.MethodGet, Path: "/admin/security/encryption-config", Exempt: true},
		{Method: http.MethodPost, Path: "/admin/security/encryption-session", Exempt: true},
	}
	prefixes := encryptionConfig.WithDefaults().ProtectedPrefixes
	jsonMethods := []string{http.MethodPost, http.MethodPut, http.MethodPatch}
	queryMethods := []string{http.MethodGet, http.MethodDelete}
	for _, prefix := range prefixes {
		for _, method := range jsonMethods {
			rules = append(rules, reqenc.Rule{
				Method: method, Path: prefix, PathPrefix: true, Location: reqenc.LocationJSON,
			})
		}
		for _, method := range queryMethods {
			rules = append(rules, reqenc.Rule{
				Method: method, Path: prefix, PathPrefix: true, Location: reqenc.LocationQuery,
			})
		}
	}
	// REQUIRED protects every supported admin API method. These catch-all rules
	// are ignored in OPTIONAL mode, where only the prefixes above are protected.
	for _, method := range jsonMethods {
		rules = append(rules, reqenc.Rule{
			Method: method, Path: "/admin/", PathPrefix: true,
			Location: reqenc.LocationJSON, RequiredOnly: true,
		})
	}
	for _, method := range queryMethods {
		rules = append(rules, reqenc.Rule{
			Method: method, Path: "/admin/", PathPrefix: true,
			Location: reqenc.LocationQuery, RequiredOnly: true,
		})
	}
	return reqenc.NewRegistry(rules...)
}
