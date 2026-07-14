package utils

import (
	"fmt"
	"sort"
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

// StockHolidayRegionCodes maps iTick stock regions to the country/region code
// accepted by /symbol/v2/holidays. SZ and SH share the mainland China calendar.
var StockHolidayRegionCodes = map[string]string{
	"HK": "HK", "SZ": "CN", "SH": "CN", "US": "US", "SG": "SG",
	"JP": "JP", "TW": "TW", "IN": "IN", "TH": "TH", "DE": "DE",
	"MX": "MX", "MY": "MY", "TR": "TR", "ES": "ES", "NL": "NL",
	"GB": "GB", "ID": "ID", "VN": "VN",
}

func StockHolidayCode(market string) (string, bool) {
	code, ok := StockHolidayRegionCodes[strings.ToUpper(strings.TrimSpace(market))]
	return code, ok
}

func StockHolidayMarketsByCode() map[string][]string {
	out := make(map[string][]string)
	for _, market := range KlineCategoryRegions["stock"] {
		if code, ok := StockHolidayCode(market); ok {
			out[code] = append(out[code], market)
		}
	}
	for code := range out {
		sort.Strings(out[code])
	}
	return out
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
