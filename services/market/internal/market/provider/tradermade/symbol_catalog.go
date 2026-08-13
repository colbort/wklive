package tradermade

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
)

type symbolCatalogKind string

const (
	restSymbolCatalog   symbolCatalogKind = "rest"
	streamSymbolCatalog symbolCatalogKind = "stream"
)

// SymbolCatalog is an authoritative TraderMade symbol directory. REST and
// streaming use different upstream directories, so each transport owns a
// separate catalog instead of assuming that a REST symbol is subscribable.
type SymbolCatalog struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	kind       symbolCatalogKind

	loadMu  sync.Mutex
	mu      sync.RWMutex
	loaded  bool
	forward map[string]string
	reverse map[string]string
}

type traderMadeDirectoryResponse struct {
	AvailableCurrencies json.RawMessage `json:"available_currencies"`
	AvailableCFDs       json.RawMessage `json:"available_cfds"`
	Code                int             `json:"code"`
	Message             string          `json:"message"`
	Error               string          `json:"error"`
}

func newRESTSymbolCatalog(baseURL, apiKey string, httpClient *http.Client) *SymbolCatalog {
	return newSymbolCatalog(baseURL, apiKey, httpClient, restSymbolCatalog)
}

func newStreamSymbolCatalog(baseURL, apiKey string, httpClient *http.Client) *SymbolCatalog {
	return newSymbolCatalog(baseURL, apiKey, httpClient, streamSymbolCatalog)
}

func newSymbolCatalog(baseURL, apiKey string, httpClient *http.Client, kind symbolCatalogKind) *SymbolCatalog {
	return &SymbolCatalog{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: httpClient,
		kind:       kind,
		forward:    make(map[string]string),
		reverse:    make(map[string]string),
	}
}

func (c *SymbolCatalog) Load(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("TraderMade symbol catalog is nil")
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

	forward := make(map[string]string)
	reverse := make(map[string]string)
	var err error
	switch c.kind {
	case restSymbolCatalog:
		err = c.loadREST(ctx, forward, reverse)
	case streamSymbolCatalog:
		err = c.loadStream(ctx, forward, reverse)
	default:
		err = fmt.Errorf("unsupported TraderMade symbol catalog kind %q", c.kind)
	}
	if err != nil {
		return err
	}
	if len(forward) == 0 {
		return fmt.Errorf("TraderMade %s symbol catalog is empty", c.kind)
	}
	c.mu.Lock()
	c.forward = forward
	c.reverse = reverse
	c.loaded = true
	c.mu.Unlock()
	return nil
}

func (c *SymbolCatalog) loadREST(ctx context.Context, forward, reverse map[string]string) error {
	currencies, err := c.fetchStringDirectory(ctx, "/live_currencies_list", "available_currencies")
	if err != nil {
		return err
	}
	// TraderMade REST publishes currency codes rather than all pair
	// permutations. Build pairs only from codes confirmed by that directory.
	for _, base := range currencies {
		for _, quote := range currencies {
			if base == quote {
				continue
			}
			addSymbolMapping(forward, reverse, base+quote, base+quote)
		}
	}
	// CFD access is plan-dependent. A rejected CFD directory must not disable
	// otherwise valid forex REST data.
	if cfds, fetchErr := c.fetchStringDirectory(ctx, "/cfd_list", "available_cfds"); fetchErr == nil {
		for _, symbol := range cfds {
			addSymbolMapping(forward, reverse, symbol, symbol)
		}
	}
	return nil
}

func (c *SymbolCatalog) loadStream(ctx context.Context, forward, reverse map[string]string) error {
	streamSymbols, err := c.fetchStringDirectory(ctx, "/streaming_currencies_list", "available_currencies")
	if err != nil {
		return err
	}
	exact := make(map[string]string, len(streamSymbols))
	for _, symbol := range streamSymbols {
		key := canonicalSymbol(symbol)
		if key == "" {
			continue
		}
		exact[key] = symbol
		addSymbolMapping(forward, reverse, key, symbol)
	}
	// Some TraderMade streaming plans expose CFDs with a USD suffix. Create an
	// internal CFD alias only when that exact suffixed symbol was returned by
	// the streaming directory; never infer an unlisted symbol.
	if cfds, fetchErr := c.fetchStringDirectory(ctx, "/cfd_list", "available_cfds"); fetchErr == nil {
		for _, cfd := range cfds {
			key := canonicalSymbol(cfd)
			if upstream, ok := exact[key]; ok {
				addSymbolMapping(forward, reverse, key, upstream)
				continue
			}
			if upstream, ok := exact[key+"USD"]; ok {
				addSymbolMapping(forward, reverse, key, upstream)
			}
		}
	}
	return nil
}

func (c *SymbolCatalog) Resolve(value string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("TraderMade symbol catalog is nil")
	}
	key := canonicalSymbol(value)
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded {
		return "", fmt.Errorf("TraderMade %s symbol catalog is not loaded", c.kind)
	}
	upstream, ok := c.forward[key]
	if !ok {
		return "", fmt.Errorf("TraderMade %s symbol is unsupported: %s", c.kind, key)
	}
	return upstream, nil
}

func (c *SymbolCatalog) Internal(value string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("TraderMade symbol catalog is nil")
	}
	key := canonicalSymbol(value)
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded {
		return "", fmt.Errorf("TraderMade %s symbol catalog is not loaded", c.kind)
	}
	internal, ok := c.reverse[key]
	if !ok {
		return "", fmt.Errorf("TraderMade %s upstream symbol is unsupported: %s", c.kind, key)
	}
	return internal, nil
}

func (c *SymbolCatalog) Count() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.forward)
}

func (c *SymbolCatalog) Loaded() bool {
	if c == nil {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}

func (c *SymbolCatalog) fetchStringDirectory(ctx context.Context, path, field string) ([]string, error) {
	if c.baseURL == "" || c.apiKey == "" || c.httpClient == nil {
		return nil, fmt.Errorf("TraderMade %s symbol catalog is not configured", c.kind)
	}
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("api_key", c.apiKey)
	endpoint.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TraderMade %s catalog request failed: %s", c.kind, sanitizedRequestError(err, c.apiKey))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody := strings.ReplaceAll(strings.TrimSpace(string(body)), c.apiKey, "[REDACTED]")
		return nil, fmt.Errorf("TraderMade %s catalog http status=%d body=%q", c.kind, resp.StatusCode, responseBody)
	}
	var response traderMadeDirectoryResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}
	if response.Code != 0 && response.Code != http.StatusOK {
		return nil, fmt.Errorf("TraderMade %s catalog rejected: code=%d message=%s", c.kind, response.Code, firstNonEmpty(response.Message, response.Error))
	}
	var raw json.RawMessage
	switch field {
	case "available_currencies":
		raw = response.AvailableCurrencies
	case "available_cfds":
		raw = response.AvailableCFDs
	default:
		return nil, fmt.Errorf("unsupported TraderMade directory field %q", field)
	}
	values, err := decodeStringDirectory(raw)
	if err != nil {
		return nil, fmt.Errorf("decode TraderMade %s catalog %s: %w", c.kind, path, err)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("TraderMade %s catalog %s is empty", c.kind, path)
	}
	return values, nil
}

func decodeStringDirectory(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, fmt.Errorf("directory field is missing")
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		return normalizeDirectoryValues(list), nil
	}
	var mapping map[string]json.RawMessage
	if err := json.Unmarshal(raw, &mapping); err != nil {
		return nil, err
	}
	list = make([]string, 0, len(mapping))
	for key := range mapping {
		list = append(list, key)
	}
	return normalizeDirectoryValues(list), nil
}

func normalizeDirectoryValues(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func addSymbolMapping(forward, reverse map[string]string, internal, upstream string) {
	internal = canonicalSymbol(internal)
	upstream = strings.ToUpper(strings.TrimSpace(upstream))
	upstreamKey := canonicalSymbol(upstream)
	if internal == "" || upstreamKey == "" {
		return
	}
	forward[internal] = upstream
	// Preserve the first (exact directory) reverse mapping. Aliases may share
	// one upstream symbol and cannot safely replace its canonical identity.
	if _, exists := reverse[upstreamKey]; !exists {
		reverse[upstreamKey] = internal
	}
}
