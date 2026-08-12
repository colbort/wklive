package tradermade

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type restResponse struct {
	Endpoint      string      `json:"endpoint"`
	Quotes        []restQuote `json:"quotes"`
	RequestedTime string      `json:"requested_time"`
	Timestamp     int64       `json:"timestamp"`
	Code          int         `json:"code"`
	Message       string      `json:"message"`
	Error         string      `json:"error"`
}

type restQuote struct {
	Ask           json.Number `json:"ask"`
	Bid           json.Number `json:"bid"`
	Mid           json.Number `json:"mid"`
	BaseCurrency  string      `json:"base_currency"`
	QuoteCurrency string      `json:"quote_currency"`
	Instrument    string      `json:"instrument"`
}

func decodeRESTResponse(data []byte) (restResponse, error) {
	var response restResponse
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return restResponse{}, err
	}
	if response.Code != 0 && response.Code != 200 {
		return restResponse{}, fmt.Errorf("TraderMade REST rejected: code=%d message=%s", response.Code, firstNonEmpty(response.Message, response.Error))
	}
	if len(response.Quotes) == 0 {
		return restResponse{}, fmt.Errorf("TraderMade REST returned no quotes: %s", firstNonEmpty(response.Message, response.Error))
	}
	return response, nil
}

func (q restQuote) symbol() string {
	if q.Instrument != "" {
		return canonicalSymbol(q.Instrument)
	}
	return canonicalSymbol(q.BaseCurrency + q.QuoteCurrency)
}

func (q restQuote) midpoint() (string, float64, error) {
	if text := numberText(q.Mid); text != "" {
		value, err := decimal.NewFromString(text)
		if err != nil {
			return "", 0, err
		}
		floatValue, _ := value.Float64()
		return text, floatValue, nil
	}
	bid, err := decimal.NewFromString(numberText(q.Bid))
	if err != nil {
		return "", 0, fmt.Errorf("invalid bid: %w", err)
	}
	ask, err := decimal.NewFromString(numberText(q.Ask))
	if err != nil {
		return "", 0, fmt.Errorf("invalid ask: %w", err)
	}
	mid := bid.Add(ask).Div(decimal.NewFromInt(2))
	floatValue, _ := mid.Float64()
	return mid.String(), floatValue, nil
}

type wsEnvelope struct {
	Type          string            `json:"type"`
	MessageType   string            `json:"t"`
	Reason        string            `json:"reason"`
	SymbolLimit   int               `json:"symbol_limit"`
	TraderLadder  bool              `json:"trader_ladder"`
	Accepted      []string          `json:"accepted"`
	Denied        []string          `json:"denied"`
	DeniedReasons map[string]string `json:"denied_reasons"`
	Invalid       []string          `json:"invalid"`
	Ask           string            `json:"a"`
	AskVolume     string            `json:"av"`
	Bid           string            `json:"b"`
	BidVolume     string            `json:"bv"`
	Mid           string            `json:"m"`
	Symbol        string            `json:"s"`
	Timestamp     string            `json:"ts"`
	Ladder        *wsLadder         `json:"ladder"`
}

type wsLadder struct {
	Asks [][]string `json:"a"`
	Bids [][]string `json:"b"`
}

func (m wsEnvelope) midpoint() (string, float64, error) {
	if strings.TrimSpace(m.Mid) != "" {
		value, err := decimal.NewFromString(strings.TrimSpace(m.Mid))
		if err != nil {
			return "", 0, err
		}
		floatValue, _ := value.Float64()
		return strings.TrimSpace(m.Mid), floatValue, nil
	}
	bid, err := decimal.NewFromString(strings.TrimSpace(m.Bid))
	if err != nil {
		return "", 0, fmt.Errorf("invalid bid: %w", err)
	}
	ask, err := decimal.NewFromString(strings.TrimSpace(m.Ask))
	if err != nil {
		return "", 0, fmt.Errorf("invalid ask: %w", err)
	}
	mid := bid.Add(ask).Div(decimal.NewFromInt(2))
	floatValue, _ := mid.Float64()
	return mid.String(), floatValue, nil
}

func parseWSTimestamp(value string) (int64, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"20060102-15:04:05.000", "20060102-15:04:05"} {
		// The wire timestamp has no offset; normalize the provider server time as UTC.
		parsed, err := time.ParseInLocation(layout, value, time.UTC)
		if err == nil {
			return parsed.UnixMilli(), nil
		}
	}
	return 0, fmt.Errorf("unsupported TraderMade timestamp %q", value)
}

func canonicalSymbol(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	return strings.NewReplacer("/", "", "-", "", "_", "", ":", "").Replace(value)
}

func numberText(value json.Number) string {
	return strings.TrimSpace(value.String())
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return "unknown error"
}
