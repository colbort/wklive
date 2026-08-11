package utils

import "strings"

// stockMarketQuoteCurrencies contains the default trading currency for each
// iTick stock region. Individual products may still be overridden in the
// admin UI when an exchange exposes a counter in another currency.
var stockMarketQuoteCurrencies = map[string]string{
	"HK": "HKD",
	"SZ": "CNY",
	"SH": "CNY",
	"US": "USD",
	"SG": "SGD",
	"JP": "JPY",
	"TW": "TWD",
	"IN": "INR",
	"TH": "THB",
	"DE": "EUR",
	"MX": "MXN",
	"MY": "MYR",
	"TR": "TRY",
	"ES": "EUR",
	"NL": "EUR",
	"GB": "GBP",
	"ID": "IDR",
	"VN": "VND",
}

func StockMarketQuoteCurrency(market string) (string, bool) {
	currency, ok := stockMarketQuoteCurrencies[strings.ToUpper(strings.TrimSpace(market))]
	return currency, ok
}

// DefaultProductAssets returns assets that can be safely inferred from iTick's
// product-list response. That response does not include its stock/info fcc
// field, so stocks use the region currency as a default and remain editable.
func DefaultProductAssets(categoryCode, market, symbol string) (string, string) {
	if NormalizeCategory(categoryCode) != "stock" {
		return "", ""
	}

	quote, _ := StockMarketQuoteCurrency(market)
	return strings.ToUpper(strings.TrimSpace(symbol)), quote
}
