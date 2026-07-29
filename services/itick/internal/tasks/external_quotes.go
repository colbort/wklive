package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"wklive/services/itick/internal/config"
	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/svc"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	externalQuoteAdapterBinanceSpot    = "BINANCE_SPOT_AGG_TRADES"
	externalQuoteAdapterBinanceFutures = "BINANCE_FUTURES_AGG_TRADES"
	externalQuoteAdapterOKXSpot        = "OKX_SPOT_TICKER"
	externalQuoteAdapterBybitSpot      = "BYBIT_SPOT_RECENT_TRADE"

	defaultExternalQuoteInterval      = time.Second
	defaultExternalQuoteTimeout       = 3 * time.Second
	defaultExternalQuoteMaxSourceAge  = 30 * time.Second
	defaultExternalQuoteMaxFutureSkew = 5 * time.Second
	externalQuoteErrorLogInterval     = 30 * time.Second
	maxExternalQuoteResponseBytes     = 1 << 20
)

type externalQuoteSource struct {
	config.ExternalQuoteSourceConf
	interval      time.Duration
	timeout       time.Duration
	maxSourceAge  time.Duration
	maxFutureSkew time.Duration
}

type externalQuoteValue struct {
	price           string
	sourceTimestamp int64
}

type externalQuoteRunner struct {
	client  *http.Client
	sources []externalQuoteSource
	handler func(context.Context, types.ClientMessage, *types.QuotePayload) error
}

func StartExternalQuoteSources(ctx context.Context, svcCtx *svc.ServiceContext) error {
	if svcCtx == nil {
		return errors.New("external quote service context is nil")
	}
	runner, err := newExternalQuoteRunner(
		svcCtx.Config.ExternalQuotes,
		svcCtx.AuthoritativeQuoteHandler,
	)
	if err != nil {
		return err
	}
	if len(runner.sources) == 0 {
		return nil
	}
	for _, source := range runner.sources {
		source := source
		go runner.run(ctx, source)
	}
	logx.Infof(
		"external public quote sources started sources=%d independent_providers=%d",
		len(runner.sources),
		distinctExternalQuoteProviders(runner.sources),
	)
	return nil
}

func newExternalQuoteRunner(
	configs []config.ExternalQuoteSourceConf,
	handler func(context.Context, types.ClientMessage, *types.QuotePayload) error,
) (*externalQuoteRunner, error) {
	sources := make([]externalQuoteSource, 0, len(configs))
	authorities := make(map[string]struct{}, len(configs))
	for _, sourceConfig := range configs {
		if !sourceConfig.Enabled {
			continue
		}
		source, err := normalizeExternalQuoteSource(sourceConfig)
		if err != nil {
			return nil, err
		}
		if _, exists := authorities[source.Authority]; exists {
			return nil, fmt.Errorf("duplicate external quote authority: %s", source.Authority)
		}
		authorities[source.Authority] = struct{}{}
		sources = append(sources, source)
	}
	if len(sources) == 0 {
		return &externalQuoteRunner{}, nil
	}
	if handler == nil {
		return nil, errors.New("external quote authoritative handler is nil")
	}
	if distinctExternalQuoteProviders(sources) < 3 {
		return nil, errors.New("external quote configuration requires at least three independent providers")
	}
	return &externalQuoteRunner{
		client:  &http.Client{},
		sources: sources,
		handler: handler,
	}, nil
}

func normalizeExternalQuoteSource(source config.ExternalQuoteSourceConf) (externalQuoteSource, error) {
	source.Authority = strings.ToLower(strings.TrimSpace(source.Authority))
	source.ProviderCode = strings.ToUpper(strings.TrimSpace(source.ProviderCode))
	source.Adapter = strings.ToUpper(strings.TrimSpace(source.Adapter))
	source.BaseURL = strings.TrimSpace(source.BaseURL)
	source.CategoryCode = strings.ToLower(strings.TrimSpace(source.CategoryCode))
	source.Market = strings.ToUpper(strings.TrimSpace(source.Market))
	source.Symbol = strings.ToUpper(strings.TrimSpace(source.Symbol))
	source.UpstreamSymbol = strings.ToUpper(strings.TrimSpace(source.UpstreamSymbol))
	if !validExternalQuoteToken(source.Authority, false) ||
		!validExternalQuoteToken(source.ProviderCode, true) ||
		source.CategoryCode == "" || source.Market == "" ||
		source.Symbol == "" || source.UpstreamSymbol == "" {
		return externalQuoteSource{}, fmt.Errorf("invalid external quote identity: %s", source.Authority)
	}
	if err := validateExternalQuoteEndpoint(source.Adapter, source.BaseURL); err != nil {
		return externalQuoteSource{}, fmt.Errorf("external quote %s: %w", source.Authority, err)
	}
	interval := time.Duration(source.IntervalMs) * time.Millisecond
	if interval == 0 {
		interval = defaultExternalQuoteInterval
	}
	if interval < 500*time.Millisecond || interval > time.Minute {
		return externalQuoteSource{}, fmt.Errorf("external quote %s interval must be within 500ms and 60s", source.Authority)
	}
	timeout := time.Duration(source.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = defaultExternalQuoteTimeout
	}
	if timeout < 100*time.Millisecond || timeout > 10*time.Second {
		return externalQuoteSource{}, fmt.Errorf("external quote %s timeout must be within 100ms and 10s", source.Authority)
	}
	maxSourceAge := time.Duration(source.MaxSourceAgeMs) * time.Millisecond
	if maxSourceAge == 0 {
		maxSourceAge = defaultExternalQuoteMaxSourceAge
	}
	if maxSourceAge < time.Second || maxSourceAge > 5*time.Minute {
		return externalQuoteSource{}, fmt.Errorf("external quote %s max source age must be within 1s and 5m", source.Authority)
	}
	maxFutureSkew := time.Duration(source.MaxFutureSkewMs) * time.Millisecond
	if maxFutureSkew == 0 {
		maxFutureSkew = defaultExternalQuoteMaxFutureSkew
	}
	if maxFutureSkew < 0 || maxFutureSkew > 30*time.Second {
		return externalQuoteSource{}, fmt.Errorf("external quote %s max future skew must be within 0 and 30s", source.Authority)
	}
	return externalQuoteSource{
		ExternalQuoteSourceConf: source,
		interval:                interval,
		timeout:                 timeout,
		maxSourceAge:            maxSourceAge,
		maxFutureSkew:           maxFutureSkew,
	}, nil
}

func distinctExternalQuoteProviders(sources []externalQuoteSource) int {
	providers := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		providers[source.ProviderCode] = struct{}{}
	}
	return len(providers)
}

func validExternalQuoteToken(value string, upper bool) bool {
	if len(value) == 0 || len(value) > 32 {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		isLetter := character >= 'a' && character <= 'z'
		if upper {
			isLetter = character >= 'A' && character <= 'Z'
		}
		if !isLetter && (character < '0' || character > '9') &&
			character != '-' && character != '_' {
			return false
		}
	}
	return value[0] != '-' && value[0] != '_' &&
		value[len(value)-1] != '-' && value[len(value)-1] != '_'
}

func validateExternalQuoteEndpoint(adapter, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("base URL must be an HTTPS endpoint without credentials, query, or fragment")
	}
	expected := map[string]struct {
		host string
		path string
	}{
		externalQuoteAdapterBinanceSpot: {
			host: "api.binance.com",
			path: "/api/v3/aggTrades",
		},
		externalQuoteAdapterBinanceFutures: {
			host: "fapi.binance.com",
			path: "/fapi/v1/aggTrades",
		},
		externalQuoteAdapterOKXSpot: {
			host: "www.okx.com",
			path: "/api/v5/market/ticker",
		},
		externalQuoteAdapterBybitSpot: {
			host: "api.bybit.com",
			path: "/v5/market/recent-trade",
		},
	}
	endpoint, ok := expected[adapter]
	if !ok {
		return fmt.Errorf("unsupported adapter: %s", adapter)
	}
	if !strings.EqualFold(parsed.Host, endpoint.host) || parsed.Path != endpoint.path {
		return fmt.Errorf("adapter %s requires https://%s%s", adapter, endpoint.host, endpoint.path)
	}
	return nil
}

func (runner *externalQuoteRunner) run(ctx context.Context, source externalQuoteSource) {
	ticker := time.NewTicker(source.interval)
	defer ticker.Stop()
	var lastErrorLog time.Time
	unhealthy := false
	for {
		err := runner.fetchAndPublish(ctx, source)
		if err != nil && ctx.Err() == nil {
			now := time.Now()
			if !unhealthy || now.Sub(lastErrorLog) >= externalQuoteErrorLogInterval {
				logx.Errorf(
					"external public quote failed authority=%s provider=%s adapter=%s err=%v",
					source.Authority, source.ProviderCode, source.Adapter, err,
				)
				lastErrorLog = now
			}
			unhealthy = true
		} else if err == nil && unhealthy {
			logx.Infof(
				"external public quote recovered authority=%s provider=%s",
				source.Authority, source.ProviderCode,
			)
			unhealthy = false
			lastErrorLog = time.Time{}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (runner *externalQuoteRunner) fetchAndPublish(
	ctx context.Context,
	source externalQuoteSource,
) error {
	requestURL, err := externalQuoteURL(source)
	if err != nil {
		return err
	}
	requestCtx, cancel := context.WithTimeout(ctx, source.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "wklive-price-source/1.0")
	response, err := runner.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, maxExternalQuoteResponseBytes+1))
	if err != nil {
		return err
	}
	if len(raw) > maxExternalQuoteResponseBytes {
		return errors.New("external quote response exceeds 1 MiB")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("external quote HTTP status %d", response.StatusCode)
	}
	value, err := parseExternalQuote(source.Adapter, source.UpstreamSymbol, raw)
	if err != nil {
		return err
	}
	now := time.Now()
	sourceTime := time.UnixMilli(value.sourceTimestamp)
	if now.Sub(sourceTime) > source.maxSourceAge {
		return fmt.Errorf("external quote is stale by %s", now.Sub(sourceTime).Round(time.Millisecond))
	}
	if sourceTime.Sub(now) > source.maxFutureSkew {
		return fmt.Errorf("external quote is ahead by %s", sourceTime.Sub(now).Round(time.Millisecond))
	}
	message := types.ClientMessage{
		Topic:        types.TopicQuote,
		CategoryCode: source.CategoryCode,
		Market:       source.Market,
		Symbol:       source.Symbol,
	}
	payload := &types.QuotePayload{
		LastPriceText: value.price,
		Ts:            value.sourceTimestamp,
		Authority:     source.Authority,
	}
	if err = runner.handler(requestCtx, message, payload); err != nil {
		return fmt.Errorf("archive external quote: %w", err)
	}
	return nil
}

func externalQuoteURL(source externalQuoteSource) (string, error) {
	parsed, err := url.Parse(source.BaseURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	switch source.Adapter {
	case externalQuoteAdapterBinanceSpot, externalQuoteAdapterBinanceFutures:
		query.Set("symbol", source.UpstreamSymbol)
		query.Set("limit", "1")
	case externalQuoteAdapterOKXSpot:
		query.Set("instId", source.UpstreamSymbol)
	case externalQuoteAdapterBybitSpot:
		query.Set("category", "spot")
		query.Set("symbol", source.UpstreamSymbol)
		query.Set("limit", "1")
	default:
		return "", fmt.Errorf("unsupported external quote adapter: %s", source.Adapter)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func parseExternalQuote(adapter, expectedSymbol string, raw []byte) (externalQuoteValue, error) {
	var value externalQuoteValue
	switch adapter {
	case externalQuoteAdapterBinanceSpot, externalQuoteAdapterBinanceFutures:
		var response []struct {
			Symbol    string `json:"s"`
			Price     string `json:"p"`
			TradeTime int64  `json:"T"`
		}
		if err := json.Unmarshal(raw, &response); err != nil {
			return value, err
		}
		if len(response) != 1 ||
			response[0].Symbol != "" && !strings.EqualFold(response[0].Symbol, expectedSymbol) {
			return value, errors.New("unexpected Binance aggregate trade response")
		}
		value = externalQuoteValue{
			price:           response[0].Price,
			sourceTimestamp: response[0].TradeTime,
		}
	case externalQuoteAdapterOKXSpot:
		var response struct {
			Code string `json:"code"`
			Data []struct {
				Instrument string `json:"instId"`
				Last       string `json:"last"`
				Timestamp  string `json:"ts"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &response); err != nil {
			return value, err
		}
		if response.Code != "0" || len(response.Data) != 1 ||
			!strings.EqualFold(response.Data[0].Instrument, expectedSymbol) {
			return value, errors.New("unexpected OKX ticker response")
		}
		timestamp, err := strconv.ParseInt(response.Data[0].Timestamp, 10, 64)
		if err != nil {
			return value, errors.New("invalid OKX ticker timestamp")
		}
		value = externalQuoteValue{
			price:           response.Data[0].Last,
			sourceTimestamp: timestamp,
		}
	case externalQuoteAdapterBybitSpot:
		var response struct {
			ReturnCode int `json:"retCode"`
			Result     struct {
				List []struct {
					Symbol    string `json:"symbol"`
					Price     string `json:"price"`
					TradeTime string `json:"time"`
				} `json:"list"`
			} `json:"result"`
		}
		if err := json.Unmarshal(raw, &response); err != nil {
			return value, err
		}
		if response.ReturnCode != 0 || len(response.Result.List) != 1 ||
			!strings.EqualFold(response.Result.List[0].Symbol, expectedSymbol) {
			return value, errors.New("unexpected Bybit recent trade response")
		}
		timestamp, err := strconv.ParseInt(response.Result.List[0].TradeTime, 10, 64)
		if err != nil {
			return value, errors.New("invalid Bybit trade timestamp")
		}
		value = externalQuoteValue{
			price:           response.Result.List[0].Price,
			sourceTimestamp: timestamp,
		}
	default:
		return value, fmt.Errorf("unsupported external quote adapter: %s", adapter)
	}
	value.price = strings.TrimSpace(value.price)
	price, err := decimal.NewFromString(value.price)
	if err != nil || !price.IsPositive() || value.sourceTimestamp <= 0 {
		return externalQuoteValue{}, errors.New("external quote price or timestamp is invalid")
	}
	return value, nil
}
