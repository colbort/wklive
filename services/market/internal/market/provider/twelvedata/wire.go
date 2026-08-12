package twelvedata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type restQuote struct {
	Code         int    `json:"code"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	Symbol       string `json:"symbol"`
	Timestamp    int64  `json:"timestamp"`
	Close        string `json:"close"`
	IsMarketOpen *bool  `json:"is_market_open"`
}

func decodeRESTQuote(data []byte) (restQuote, error) {
	var response restQuote
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return restQuote{}, err
	}
	if strings.EqualFold(response.Status, "error") || response.Code >= 400 {
		return restQuote{}, fmt.Errorf("Twelve Data REST rejected: code=%d message=%s", response.Code, firstNonEmpty(response.Message, response.Status))
	}
	if canonicalSymbol(response.Symbol) == "" || strings.TrimSpace(response.Close) == "" || response.Timestamp <= 0 {
		return restQuote{}, fmt.Errorf("Twelve Data REST returned incomplete quote for %q", response.Symbol)
	}
	if response.IsMarketOpen != nil && !*response.IsMarketOpen {
		return restQuote{}, fmt.Errorf("Twelve Data reports market closed for %s", response.Symbol)
	}
	return response, nil
}

func (q restQuote) price() (string, float64, error) {
	text := strings.TrimSpace(q.Close)
	value, err := decimal.NewFromString(text)
	if err != nil || !value.IsPositive() {
		return "", 0, fmt.Errorf("invalid Twelve Data close price %q", text)
	}
	floatValue, _ := value.Float64()
	return text, floatValue, nil
}

type wsEnvelope struct {
	Event     string          `json:"event"`
	Status    string          `json:"status"`
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Symbol    string          `json:"symbol"`
	Timestamp int64           `json:"timestamp"`
	Price     json.RawMessage `json:"price"`
	DayVolume json.RawMessage `json:"day_volume"`
	Success   []wsStatusItem  `json:"success"`
	Fails     []wsStatusItem  `json:"fails"`
}

type wsStatusItem struct {
	Symbol  string `json:"symbol"`
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func rawDecimal(raw json.RawMessage) (string, float64, error) {
	text := strings.TrimSpace(string(raw))
	if unquoted, err := strconv.Unquote(text); err == nil {
		text = strings.TrimSpace(unquoted)
	}
	value, err := decimal.NewFromString(text)
	if err != nil || !value.IsPositive() {
		return "", 0, fmt.Errorf("invalid price %q", text)
	}
	floatValue, _ := value.Float64()
	return text, floatValue, nil
}

func rawFloat(raw json.RawMessage) float64 {
	text := strings.TrimSpace(string(raw))
	if unquoted, err := strconv.Unquote(text); err == nil {
		text = strings.TrimSpace(unquoted)
	}
	value, _ := strconv.ParseFloat(text, 64)
	return value
}

func canonicalSymbol(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	return strings.NewReplacer("/", "", "-", "", "_", "", ":", "").Replace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return "unknown error"
}
