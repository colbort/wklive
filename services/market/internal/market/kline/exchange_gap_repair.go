package kline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"wklive/services/market/models"
)

const (
	binanceSpotKlinesURL              = "https://api.binance.com/api/v3/klines"
	binanceSpotKlinePageSize          = 1000
	maxExchangeKlineResponseBodyBytes = 8 << 20
)

var binanceSpotKlineHTTPClient = &http.Client{Timeout: 15 * time.Second}

type gapRange struct {
	start int64
	end   int64
}

func supportsExchangeKlineFallback(job *GapRepairJob) bool {
	if job == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(job.Category), "crypto") &&
		strings.EqualFold(strings.TrimSpace(job.Market), "BA") &&
		strings.EqualFold(strings.TrimSpace(job.Exchange), "Binance")
}

func (s *GapRepairService) repairMissingFromExchange(job *GapRepairJob) (int, error) {
	ranges, err := s.findMissingTradingRanges(job)
	if err != nil {
		return 0, err
	}
	if len(ranges) == 0 {
		return 0, nil
	}

	all := make([]*models.CoinKline, 0)
	for _, missing := range ranges {
		list, fetchErr := fetchBinanceSpotKlines(s.ctx, binanceSpotKlineHTTPClient, job, missing.start, missing.end)
		if fetchErr != nil {
			return len(all), fetchErr
		}
		all = append(all, list...)
	}
	if len(all) == 0 {
		return 0, fmt.Errorf("Binance returned no K lines for %d missing ranges", len(ranges))
	}
	if err = s.worker.bulkUpsertKlines(job.Category, "1m", all); err != nil {
		return 0, err
	}
	if s.svcCtx.RebuildHistoricalKlines != nil {
		if err = s.svcCtx.RebuildHistoricalKlines(all); err != nil {
			return 0, fmt.Errorf("rebuild derived klines: %w", err)
		}
	}
	return len(all), nil
}

func (s *GapRepairService) findMissingTradingRanges(job *GapRepairJob) ([]gapRange, error) {
	model := s.svcCtx.Factory.New(job.Category, "1m")
	if model == nil {
		return nil, fmt.Errorf("invalid 1m model for category=%s", job.Category)
	}
	list, err := model.FindRangeByMarketSymbol(s.ctx, job.Market, job.Symbol, job.StartTs, job.EndTs+minuteMs)
	if err != nil {
		return nil, err
	}
	present := make(map[int64]struct{}, len(list))
	for _, bar := range list {
		if bar != nil {
			present[bar.Ts] = struct{}{}
		}
	}

	ranges := make([]gapRange, 0)
	var current *gapRange
	for ts := job.StartTs; ts <= job.EndTs; ts += minuteMs {
		trading := s.svcCtx.MarketCalendarResolver == nil ||
			s.svcCtx.MarketCalendarResolver.IsTradingMinute(s.ctx, job.Category, job.Market, job.Exchange, ts)
		_, exists := present[ts]
		if trading && !exists {
			if current == nil {
				current = &gapRange{start: ts, end: ts}
			} else {
				current.end = ts
			}
			continue
		}
		if current != nil {
			ranges = append(ranges, *current)
			current = nil
		}
	}
	if current != nil {
		ranges = append(ranges, *current)
	}
	return ranges, nil
}

func fetchBinanceSpotKlines(ctx context.Context, client *http.Client, job *GapRepairJob, startTs, endTs int64) ([]*models.CoinKline, error) {
	if client == nil {
		return nil, fmt.Errorf("Binance K line HTTP client is nil")
	}
	if startTs <= 0 || endTs < startTs {
		return nil, fmt.Errorf("invalid Binance K line range: %d-%d", startTs, endTs)
	}

	result := make([]*models.CoinKline, 0, int((endTs-startTs)/minuteMs)+1)
	for cursor := startTs; cursor <= endTs; {
		requestURL, err := buildBinanceSpotKlineURL(job.Symbol, cursor, endTs)
		if err != nil {
			return nil, err
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Accept", "application/json")
		request.Header.Set("User-Agent", "wklive-kline-gap-repair/1.0")
		response, err := client.Do(request)
		if err != nil {
			return nil, fmt.Errorf("request Binance K lines: %w", err)
		}
		raw, readErr := io.ReadAll(io.LimitReader(response.Body, maxExchangeKlineResponseBodyBytes+1))
		closeErr := response.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if len(raw) > maxExchangeKlineResponseBodyBytes {
			return nil, fmt.Errorf("Binance K line response exceeds %d bytes", maxExchangeKlineResponseBodyBytes)
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			body := strings.TrimSpace(string(raw))
			if len(body) > 512 {
				body = body[:512]
			}
			return nil, fmt.Errorf("Binance K line HTTP status=%d body=%s", response.StatusCode, body)
		}
		page, err := parseBinanceSpotKlines(raw, job, cursor, endTs)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		result = append(result, page...)
		lastTs := page[len(page)-1].Ts
		if lastTs < cursor {
			return nil, fmt.Errorf("Binance K line cursor did not advance: cursor=%d last=%d", cursor, lastTs)
		}
		cursor = lastTs + minuteMs
	}
	return result, nil
}

func buildBinanceSpotKlineURL(symbol string, startTs, endTs int64) (string, error) {
	parsed, err := url.Parse(binanceSpotKlinesURL)
	if err != nil {
		return "", err
	}
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return "", fmt.Errorf("Binance symbol is empty")
	}
	query := parsed.Query()
	query.Set("symbol", symbol)
	query.Set("interval", "1m")
	query.Set("startTime", strconv.FormatInt(startTs, 10))
	query.Set("endTime", strconv.FormatInt(endTs, 10))
	query.Set("limit", strconv.Itoa(binanceSpotKlinePageSize))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func parseBinanceSpotKlines(raw []byte, job *GapRepairJob, startTs, endTs int64) ([]*models.CoinKline, error) {
	var rows [][]json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("decode Binance K lines: %w", err)
	}
	result := make([]*models.CoinKline, 0, len(rows))
	for index, row := range rows {
		if len(row) < 8 {
			return nil, fmt.Errorf("Binance K line row %d has %d fields", index, len(row))
		}
		ts, err := parseJSONInt64(row[0])
		if err != nil {
			return nil, fmt.Errorf("Binance K line row %d timestamp: %w", index, err)
		}
		if ts < startTs || ts > endTs || ts%minuteMs != 0 {
			continue
		}
		values := make([]float64, 6)
		for field, rawValue := range []json.RawMessage{row[1], row[2], row[3], row[4], row[5], row[7]} {
			values[field], err = parseJSONFloat(rawValue)
			if err != nil {
				return nil, fmt.Errorf("Binance K line row %d field %d: %w", index, field+1, err)
			}
		}
		if values[2] > values[0] || values[2] > values[3] || values[1] < values[0] ||
			values[1] < values[3] || values[1] < values[2] || values[4] < 0 || values[5] < 0 {
			return nil, fmt.Errorf("Binance K line row %d contains invalid OHLCV values", index)
		}
		result = append(result, &models.CoinKline{
			CategoryCode:  job.Category,
			Market:        job.Market,
			Symbol:        job.Symbol,
			Interval:      "1m",
			Ts:            ts,
			Open:          values[0],
			High:          values[1],
			Low:           values[2],
			Close:         values[3],
			Volume:        values[4],
			Turnover:      values[5],
			Source:        models.KlineSourceExchangeRest,
			Revision:      time.Now().UnixMilli(),
			IsClosed:      true,
			Confirmed:     true,
			ActualCount:   1,
			ExpectedCount: 1,
		})
	}
	return result, nil
}

func parseJSONInt64(raw json.RawMessage) (int64, error) {
	value := strings.Trim(string(raw), "\"")
	return strconv.ParseInt(value, 10, 64)
}

func parseJSONFloat(raw json.RawMessage) (float64, error) {
	value := strings.Trim(string(raw), "\"")
	return strconv.ParseFloat(value, 64)
}
