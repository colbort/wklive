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

// stockMarketExchanges is the exchange set currently documented and returned
// by iTick for each supported stock region. Product-list responses do not carry
// the region on each item, so this mapping is also used as a guardrail against
// an upstream response accidentally containing instruments from other regions.
var stockMarketExchanges = map[string]map[string]struct{}{
	"HK": exchangeSet("HKEX"),
	"SZ": exchangeSet("SZSE"),
	"SH": exchangeSet("SSE"),
	"US": exchangeSet("AMEX", "CBOE", "NASDAQ", "NYSE", "OTC"),
	"SG": exchangeSet("SGX"),
	"JP": exchangeSet("FSE", "NAG", "SAPSE", "TSE"),
	"TW": exchangeSet("TPEX", "TWSE"),
	"IN": exchangeSet("BSE", "NSE"),
	"TH": exchangeSet("SET"),
	"DE": exchangeSet("FWB", "XETR"),
	"MX": exchangeSet("BIVA", "BMV"),
	"MY": exchangeSet("MYX"),
	"TR": exchangeSet("BIST"),
	"ES": exchangeSet("BME"),
	"NL": exchangeSet("EURONEXT"),
	"GB": exchangeSet("LSE", "LSIN"),
	"ID": exchangeSet("IDX"),
	"VN": exchangeSet("HNX", "HOSE", "UPCOM"),
}

func exchangeSet(exchanges ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(exchanges))
	for _, exchange := range exchanges {
		result[exchange] = struct{}{}
	}
	return result
}

// StockExchangeMatchesMarket reports whether an iTick exchange belongs to the
// requested stock region. Unknown regions and blank exchanges fail closed.
func StockExchangeMatchesMarket(market, exchange string) bool {
	exchanges, ok := stockMarketExchanges[strings.ToUpper(strings.TrimSpace(market))]
	if !ok {
		return false
	}
	_, ok = exchanges[strings.ToUpper(strings.TrimSpace(exchange))]
	return ok
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
