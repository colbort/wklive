package twelvedata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

const forexCatalogPageSize = 5000

// SymbolCatalog is the authoritative Twelve Data forex symbol directory.
// Keys use the system's separator-free form and values preserve Twelve Data's
// original symbol exactly, for example USDCNY -> USD/CNY.
type SymbolCatalog struct {
	baseURL    string
	apiKey     string
	limiter    *rate.Limiter
	httpClient *http.Client

	loadMu  sync.Mutex
	mu      sync.RWMutex
	loaded  bool
	symbols map[string]string
}

type forexPairsResponse struct {
	Count   int         `json:"count"`
	Data    []forexPair `json:"data"`
	Status  string      `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
}

type forexPair struct {
	Symbol string `json:"symbol"`
}

func newSymbolCatalog(baseURL, apiKey string, limiter *rate.Limiter, httpClient *http.Client) *SymbolCatalog {
	return &SymbolCatalog{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		limiter:    limiter,
		httpClient: httpClient,
		symbols:    make(map[string]string),
	}
}

func (c *SymbolCatalog) Load(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("Twelve Data forex symbol catalog is nil")
	}
	c.mu.RLock()
	loaded := c.loaded
	c.mu.RUnlock()
	if loaded {
		return nil
	}

	c.loadMu.Lock()
	defer c.loadMu.Unlock()
	c.mu.RLock()
	loaded = c.loaded
	c.mu.RUnlock()
	if loaded {
		return nil
	}

	symbols := make(map[string]string)
	for page := 1; page <= 100; page++ {
		response, err := c.fetchPage(ctx, page)
		if err != nil {
			return err
		}
		for _, item := range response.Data {
			upstream := strings.ToUpper(strings.TrimSpace(item.Symbol))
			key := canonicalSymbol(upstream)
			if key == "" || upstream == "" {
				continue
			}
			symbols[key] = upstream
		}
		if len(response.Data) < forexCatalogPageSize {
			break
		}
	}
	if len(symbols) == 0 {
		return fmt.Errorf("Twelve Data forex symbol catalog is empty")
	}
	c.mu.Lock()
	c.symbols = symbols
	c.loaded = true
	c.mu.Unlock()
	return nil
}

func (c *SymbolCatalog) Resolve(value string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("Twelve Data forex symbol catalog is nil")
	}
	key := canonicalSymbol(value)
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded {
		return "", fmt.Errorf("Twelve Data forex symbol catalog is not loaded")
	}
	upstream, ok := c.symbols[key]
	if !ok {
		return "", fmt.Errorf("Twelve Data forex symbol is unsupported: %s", key)
	}
	return upstream, nil
}

func (c *SymbolCatalog) Count() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.symbols)
}

func (c *SymbolCatalog) Loaded() bool {
	if c == nil {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}

func (c *SymbolCatalog) fetchPage(ctx context.Context, page int) (forexPairsResponse, error) {
	if c.baseURL == "" || c.apiKey == "" || c.httpClient == nil {
		return forexPairsResponse{}, fmt.Errorf("Twelve Data forex symbol catalog is not configured")
	}
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return forexPairsResponse{}, fmt.Errorf("wait for Twelve Data REST rate limit: %w", err)
		}
	}
	endpoint, err := url.Parse(c.baseURL + "/forex_pairs")
	if err != nil {
		return forexPairsResponse{}, err
	}
	query := endpoint.Query()
	query.Set("page", strconv.Itoa(page))
	query.Set("outputsize", strconv.Itoa(forexCatalogPageSize))
	endpoint.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return forexPairsResponse{}, err
	}
	req.Header.Set("Authorization", "apikey "+c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return forexPairsResponse{}, fmt.Errorf("Twelve Data forex catalog request failed: %s", sanitizedRequestError(err, c.apiKey))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return forexPairsResponse{}, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody := strings.ReplaceAll(strings.TrimSpace(string(body)), c.apiKey, "[REDACTED]")
		return forexPairsResponse{}, fmt.Errorf("Twelve Data forex catalog http status=%d body=%q", resp.StatusCode, responseBody)
	}
	var result forexPairsResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&result); err != nil {
		return forexPairsResponse{}, err
	}
	if strings.EqualFold(result.Status, "error") || result.Code >= 400 {
		return forexPairsResponse{}, fmt.Errorf("Twelve Data forex catalog rejected: code=%d message=%s", result.Code, firstNonEmpty(result.Message, result.Status))
	}
	return result, nil
}
