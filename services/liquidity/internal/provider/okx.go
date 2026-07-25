package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/models"

	"github.com/shopspring/decimal"
)

const defaultOKXBaseURL = "https://www.okx.com"

type OKXCredentials struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	BaseURL    string
}

type CredentialResolver interface {
	Resolve(ctx context.Context, credentialRef string) (OKXCredentials, error)
}

// EnvCredentialResolver resolves refs in the form env:OKX_MM_MAIN from:
// OKX_MM_MAIN_API_KEY, OKX_MM_MAIN_SECRET_KEY and OKX_MM_MAIN_PASSPHRASE.
// An optional OKX_MM_MAIN_BASE_URL is useful for regional endpoints or tests.
type EnvCredentialResolver struct{}

func (EnvCredentialResolver) Resolve(_ context.Context, credentialRef string) (OKXCredentials, error) {
	const prefix = "env:"
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(credentialRef)), prefix) {
		return OKXCredentials{}, fmt.Errorf("OKX credential_ref must use env:<PREFIX>")
	}
	name := strings.TrimSpace(credentialRef[len(prefix):])
	if name == "" {
		return OKXCredentials{}, fmt.Errorf("OKX credential environment prefix is required")
	}
	credentials := OKXCredentials{
		APIKey:     os.Getenv(name + "_API_KEY"),
		SecretKey:  os.Getenv(name + "_SECRET_KEY"),
		Passphrase: os.Getenv(name + "_PASSPHRASE"),
		BaseURL:    os.Getenv(name + "_BASE_URL"),
	}
	if credentials.APIKey == "" || credentials.SecretKey == "" || credentials.Passphrase == "" {
		return OKXCredentials{}, fmt.Errorf("OKX credentials are incomplete for environment prefix %s", name)
	}
	return credentials, nil
}

type OKXAdapter struct {
	enabled  bool
	client   *http.Client
	resolver CredentialResolver
	baseURL  string
	now      func() time.Time
}

func NewOKXAdapter(enabled bool, resolver CredentialResolver, baseURL string, timeout time.Duration) *OKXAdapter {
	if resolver == nil {
		resolver = EnvCredentialResolver{}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultOKXBaseURL
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &OKXAdapter{
		enabled:  enabled,
		client:   &http.Client{Timeout: timeout},
		resolver: resolver,
		baseURL:  strings.TrimRight(baseURL, "/"),
		now:      time.Now,
	}
}

type okxEnvelope struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type okxOrder struct {
	OrdID        string `json:"ordId"`
	ClOrdID      string `json:"clOrdId"`
	State        string `json:"state"`
	AccFillSz    string `json:"accFillSz"`
	AvgPx        string `json:"avgPx"`
	Fee          string `json:"fee"`
	FeeCcy       string `json:"feeCcy"`
	SCode        string `json:"sCode"`
	SMsg         string `json:"sMsg"`
	CancelSource string `json:"cancelSource"`
}

func (a *OKXAdapter) Health(ctx context.Context, provider *models.TLiquidityProvider) error {
	var data []json.RawMessage
	return a.request(ctx, provider, http.MethodGet, "/api/v5/account/config", nil, nil, &data)
}

func (a *OKXAdapter) SubmitOrder(ctx context.Context, provider *models.TLiquidityProvider, order *models.TLiquidityExternalOrder) (*OrderResult, error) {
	if order == nil || strings.TrimSpace(order.ExternalSymbol) == "" {
		return nil, fmt.Errorf("OKX external symbol is required")
	}
	body := map[string]string{
		"instId":  order.ExternalSymbol,
		"tdMode":  okxTradeMode(order.ExternalSymbol),
		"side":    okxSide(order.Side),
		"ordType": okxOrderType(order.OrderType, order.TimeInForce),
		"sz":      order.Qty.String(),
	}
	if order.ExternalClientOrderId != "" {
		body["clOrdId"] = order.ExternalClientOrderId
	}
	if order.OrderType == int64(liquidity.ExternalOrderType_EXTERNAL_ORDER_TYPE_LIMIT) {
		body["px"] = order.Price.String()
	}
	if body["tdMode"] == "cash" && body["ordType"] == "market" && order.Side == 1 {
		body["tgtCcy"] = "base_ccy"
	}
	var rows []okxOrder
	if err := a.request(ctx, provider, http.MethodPost, "/api/v5/trade/order", nil, body, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("OKX place order returned no data")
	}
	if rows[0].SCode != "" && rows[0].SCode != "0" {
		return nil, fmt.Errorf("OKX place order rejected: code=%s msg=%s", rows[0].SCode, rows[0].SMsg)
	}
	return okxOrderResult(rows[0])
}

func (a *OKXAdapter) CancelOrder(ctx context.Context, provider *models.TLiquidityProvider, order *models.TLiquidityExternalOrder) (*OrderResult, error) {
	body := okxOrderIdentity(order)
	var rows []okxOrder
	if err := a.request(ctx, provider, http.MethodPost, "/api/v5/trade/cancel-order", nil, body, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("OKX cancel order returned no data")
	}
	if rows[0].SCode != "" && rows[0].SCode != "0" {
		return nil, fmt.Errorf("OKX cancel order rejected: code=%s msg=%s", rows[0].SCode, rows[0].SMsg)
	}
	result, err := okxOrderResult(rows[0])
	if err == nil {
		result.Status = int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_CANCELING)
	}
	return result, err
}

func (a *OKXAdapter) QueryOrder(ctx context.Context, provider *models.TLiquidityProvider, order *models.TLiquidityExternalOrder) (*OrderResult, error) {
	query := url.Values{"instId": []string{order.ExternalSymbol}}
	if order.ExternalOrderId.Valid && order.ExternalOrderId.String != "" {
		query.Set("ordId", order.ExternalOrderId.String)
	} else if order.ExternalClientOrderId != "" {
		query.Set("clOrdId", order.ExternalClientOrderId)
	} else {
		return nil, fmt.Errorf("OKX order id or client order id is required")
	}
	var rows []okxOrder
	if err := a.request(ctx, provider, http.MethodGet, "/api/v5/trade/order", query, nil, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("OKX query order returned no data")
	}
	return okxOrderResult(rows[0])
}

func (a *OKXAdapter) QueryFills(ctx context.Context, provider *models.TLiquidityProvider, order *models.TLiquidityExternalOrder) ([]Fill, error) {
	query := url.Values{"instId": []string{order.ExternalSymbol}}
	if order.ExternalOrderId.Valid {
		query.Set("ordId", order.ExternalOrderId.String)
	}
	var rows []struct {
		TradeID    string `json:"tradeId"`
		Side       string `json:"side"`
		FillPx     string `json:"fillPx"`
		FillSz     string `json:"fillSz"`
		FillFee    string `json:"fillFee"`
		FillFeeCcy string `json:"fillFeeCcy"`
		ExecType   string `json:"execType"`
		Ts         string `json:"ts"`
	}
	if err := a.request(ctx, provider, http.MethodGet, "/api/v5/trade/fills", query, nil, &rows); err != nil {
		return nil, err
	}
	fills := make([]Fill, 0, len(rows))
	for _, row := range rows {
		price := decimalFromString(row.FillPx)
		qty := decimalFromString(row.FillSz)
		fills = append(fills, Fill{
			ExternalTradeID: row.TradeID,
			Side:            sideFromOKX(row.Side),
			Price:           price,
			Qty:             qty,
			Amount:          price.Mul(qty),
			FeeAmount:       decimalFromString(row.FillFee).Abs(),
			FeeAsset:        row.FillFeeCcy,
			LiquidityType:   liquidityTypeFromOKX(row.ExecType),
			TradeTime:       int64FromString(row.Ts),
			RawPayload:      marshalString(row),
		})
	}
	return fills, nil
}

func (a *OKXAdapter) SnapshotInventory(ctx context.Context, provider *models.TLiquidityProvider, config *models.TLiquiditySymbolConfig) (*Inventory, error) {
	if config == nil || strings.TrimSpace(config.ExternalSymbol) == "" {
		return nil, fmt.Errorf("OKX external symbol is required")
	}
	base, quote := okxAssets(config.ExternalSymbol)
	query := url.Values{}
	if base != "" && quote != "" {
		query.Set("ccy", base+","+quote)
	}
	var balances []struct {
		Details []struct {
			Ccy       string `json:"ccy"`
			Eq        string `json:"eq"`
			AvailBal  string `json:"availBal"`
			FrozenBal string `json:"frozenBal"`
		} `json:"details"`
	}
	if err := a.request(ctx, provider, http.MethodGet, "/api/v5/account/balance", query, nil, &balances); err != nil {
		return nil, err
	}
	inventory := &Inventory{BaseAsset: base, QuoteAsset: quote}
	for _, account := range balances {
		for _, detail := range account.Details {
			switch detail.Ccy {
			case base:
				inventory.BaseTotal = decimalFromString(detail.Eq)
				inventory.BaseAvailable = decimalFromString(detail.AvailBal)
				inventory.BaseFrozen = decimalFromString(detail.FrozenBal)
			case quote:
				inventory.QuoteTotal = decimalFromString(detail.Eq)
				inventory.QuoteAvailable = decimalFromString(detail.AvailBal)
				inventory.QuoteFrozen = decimalFromString(detail.FrozenBal)
			}
		}
	}
	if okxTradeMode(config.ExternalSymbol) != "cash" {
		var positions []struct {
			Pos string `json:"pos"`
		}
		positionQuery := url.Values{"instId": []string{config.ExternalSymbol}}
		if err := a.request(ctx, provider, http.MethodGet, "/api/v5/account/positions", positionQuery, nil, &positions); err != nil {
			return nil, err
		}
		for _, position := range positions {
			inventory.PositionQty = inventory.PositionQty.Add(decimalFromString(position.Pos))
		}
	}
	inventory.RawPayload = marshalString(balances)
	return inventory, nil
}

func (a *OKXAdapter) request(ctx context.Context, provider *models.TLiquidityProvider, method, path string, query url.Values, body any, out any) error {
	if !a.enabled {
		return fmt.Errorf("okx was not allowed")
	}
	if provider == nil {
		return fmt.Errorf("liquidity provider is required")
	}
	credentials, err := a.resolver.Resolve(ctx, provider.CredentialRef)
	if err != nil {
		return err
	}
	baseURL := a.baseURL
	if strings.TrimSpace(credentials.BaseURL) != "" {
		baseURL = strings.TrimRight(credentials.BaseURL, "/")
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var payload []byte
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode OKX request: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	timestamp := a.now().UTC().Format("2006-01-02T15:04:05.000Z")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OK-ACCESS-KEY", credentials.APIKey)
	req.Header.Set("OK-ACCESS-SIGN", okxSign(credentials.SecretKey, timestamp, method, path, string(payload)))
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
	if provider.Environment == int64(liquidity.ProviderEnvironment_PROVIDER_ENVIRONMENT_SANDBOX) {
		req.Header.Set("x-simulated-trading", "1")
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("OKX request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read OKX response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("OKX HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var envelope okxEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("decode OKX response: %w", err)
	}
	if envelope.Code != "0" {
		return fmt.Errorf("OKX API error: code=%s msg=%s", envelope.Code, envelope.Msg)
	}
	if out != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("decode OKX data: %w", err)
		}
	}
	return nil
}

func okxSign(secret, timestamp, method, path, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + strings.ToUpper(method) + path + body))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func okxOrderIdentity(order *models.TLiquidityExternalOrder) map[string]string {
	body := map[string]string{"instId": order.ExternalSymbol}
	if order.ExternalOrderId.Valid && order.ExternalOrderId.String != "" {
		body["ordId"] = order.ExternalOrderId.String
	} else {
		body["clOrdId"] = order.ExternalClientOrderId
	}
	return body
}

func okxOrderResult(row okxOrder) (*OrderResult, error) {
	status, err := statusFromOKX(row.State)
	if err != nil {
		return nil, err
	}
	return &OrderResult{
		ExternalOrderID: row.OrdID,
		Status:          status,
		FilledQty:       decimalFromString(row.AccFillSz),
		AvgPrice:        decimalFromString(row.AvgPx),
		FeeAmount:       decimalFromString(row.Fee).Abs(),
		FeeAsset:        row.FeeCcy,
		RawResponse:     marshalString(row),
	}, nil
}

func statusFromOKX(state string) (int64, error) {
	switch state {
	case "", "live":
		return int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_SUBMITTED), nil
	case "partially_filled":
		return int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_PART_FILLED), nil
	case "filled":
		return int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_FILLED), nil
	case "canceled", "mmp_canceled":
		return int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_CANCELED), nil
	default:
		return 0, fmt.Errorf("unknown OKX order state %q", state)
	}
}

func okxOrderType(orderType, timeInForce int64) string {
	if orderType == int64(liquidity.ExternalOrderType_EXTERNAL_ORDER_TYPE_MARKET) {
		return "market"
	}
	switch liquidity.ExternalTimeInForce(timeInForce) {
	case liquidity.ExternalTimeInForce_EXTERNAL_TIME_IN_FORCE_IOC:
		return "ioc"
	case liquidity.ExternalTimeInForce_EXTERNAL_TIME_IN_FORCE_FOK:
		return "fok"
	case liquidity.ExternalTimeInForce_EXTERNAL_TIME_IN_FORCE_POST_ONLY:
		return "post_only"
	default:
		return "limit"
	}
}

func okxSide(side int64) string {
	if side == 2 {
		return "sell"
	}
	return "buy"
}

func sideFromOKX(side string) int64 {
	if side == "sell" {
		return 2
	}
	return 1
}

func okxTradeMode(symbol string) string {
	if strings.HasSuffix(strings.ToUpper(symbol), "-SWAP") || len(strings.Split(symbol, "-")) > 2 {
		return "cross"
	}
	return "cash"
}

func okxAssets(symbol string) (string, string) {
	parts := strings.Split(strings.ToUpper(strings.TrimSpace(symbol)), "-")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func liquidityTypeFromOKX(execType string) int64 {
	if strings.EqualFold(execType, "M") {
		return 1
	}
	if strings.EqualFold(execType, "T") {
		return 2
	}
	return 0
}

func decimalFromString(value string) decimal.Decimal {
	result, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero
	}
	return result
}

func int64FromString(value string) int64 {
	result, _ := strconv.ParseInt(value, 10, 64)
	return result
}

func marshalString(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
