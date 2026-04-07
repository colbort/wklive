package utils

import (
	"fmt"
	"strings"
)

var KlineCategoryRegions = map[string][]string{
	"stock": {
		"HK", "SZ", "SH", "US", "SG", "JP", "TW", "IN", "TH",
		"DE", "MX", "MY", "TR", "ES", "NL", "GB", "ID", "VN",
	}, // 股票
	"forex":   {"GB"},             // 外汇
	"indices": {"GB"},             // 指数
	"crypto":  {"BA"},             // 加密货币
	"future":  {"US", "HK", "CN"}, // 期货
	"fund":    {"US"},             // 基金
}

var DefaultKlineCategories = func() []string {
	keys := make([]string, 0, len(KlineCategoryRegions))
	for k := range KlineCategoryRegions {
		keys = append(keys, k)
	}
	return keys
}()

func NormalizeCategory(category string) string {
	return strings.ToLower(strings.TrimSpace(category))
}

func IsSupportedKlineCategory(category string) bool {
	category = NormalizeCategory(category)
	_, ok := KlineCategoryRegions[category]
	return ok
}

func GetKlineCategoryRegions(category string) ([]string, error) {
	category = NormalizeCategory(category)

	regions, ok := KlineCategoryRegions[category]
	if !ok {
		return nil, fmt.Errorf("unsupported category: %s", category)
	}

	// 返回副本，避免外部误改底层切片
	out := make([]string, len(regions))
	copy(out, regions)
	return out, nil
}
