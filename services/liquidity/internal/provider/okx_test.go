package provider

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"wklive/proto/liquidity"
	"wklive/services/liquidity/models"

	"github.com/shopspring/decimal"
)

type staticCredentialResolver struct {
	credentials OKXCredentials
}

func (r staticCredentialResolver) Resolve(context.Context, string) (OKXCredentials, error) {
	return r.credentials, nil
}

func TestOKXSign(t *testing.T) {
	got := okxSign("secret", "2020-12-08T09:08:57.715Z", http.MethodGet, "/api/v5/account/balance?ccy=BTC", "")
	const want = "wpDvCwYCprcMQsQkxWJiWy+YADoQE4ep+OEKKLimMoY="
	if got != want {
		t.Fatalf("unexpected signature: got %q want %q", got, want)
	}
}

func TestOKXSubmitSpotMarketBuy(t *testing.T) {
	var received map[string]string
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("x-simulated-trading") != "1" {
			t.Error("missing simulated trading header")
		}
		if r.Header.Get("OK-ACCESS-SIGN") == "" {
			t.Error("missing signature")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		return okResponse(`{"code":"0","msg":"","data":[{"ordId":"123","clOrdId":"route-1","sCode":"0","sMsg":""}]}`), nil
	})

	adapter := NewOKXAdapter(false, staticCredentialResolver{OKXCredentials{
		APIKey: "key", SecretKey: "secret", Passphrase: "passphrase",
	}}, "", time.Second)
	adapter.client.Transport = transport
	adapter.now = func() time.Time { return time.Unix(0, 0) }
	result, err := adapter.SubmitOrder(context.Background(), &models.TLiquidityProvider{
		CredentialRef: "env:TEST",
		Environment:   int64(liquidity.ProviderEnvironment_PROVIDER_ENVIRONMENT_SANDBOX),
	}, &models.TLiquidityExternalOrder{
		ExternalSymbol:        "BTC-USDT",
		ExternalClientOrderId: "route-1",
		Side:                  1,
		OrderType:             int64(liquidity.ExternalOrderType_EXTERNAL_ORDER_TYPE_MARKET),
		Qty:                   decimal.RequireFromString("0.01"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ExternalOrderID != "123" {
		t.Fatalf("unexpected order id %q", result.ExternalOrderID)
	}
	if received["tdMode"] != "cash" || received["tgtCcy"] != "base_ccy" {
		t.Fatalf("unexpected spot market request: %#v", received)
	}
	if received["sz"] != "0.01" {
		t.Fatalf("unexpected size %q", received["sz"])
	}
}

func TestOKXQueryOrderMapping(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("ordId") != "123" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		return okResponse(`{"code":"0","msg":"","data":[{"ordId":"123","state":"partially_filled","accFillSz":"0.2","avgPx":"64000","fee":"-0.01","feeCcy":"USDT"}]}`), nil
	})

	adapter := NewOKXAdapter(false, staticCredentialResolver{OKXCredentials{
		APIKey: "key", SecretKey: "secret", Passphrase: "passphrase",
	}}, "", time.Second)
	adapter.client.Transport = transport
	result, err := adapter.QueryOrder(context.Background(), &models.TLiquidityProvider{
		CredentialRef: "env:TEST",
	}, &models.TLiquidityExternalOrder{
		ExternalSymbol:  "BTC-USDT",
		ExternalOrderId: sql.NullString{String: "123", Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_PART_FILLED) {
		t.Fatalf("unexpected status %d", result.Status)
	}
	if !result.FilledQty.Equal(decimal.RequireFromString("0.2")) ||
		!result.FeeAmount.Equal(decimal.RequireFromString("0.01")) {
		t.Fatalf("unexpected result: %#v", result)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func okResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
