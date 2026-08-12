package twelvedata

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"wklive/services/market/internal/market/provider"
)

func TestRESTFetchQuoteUsesHeaderAuthenticationAndSourceTimestamp(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.Header.Get("Authorization"); got != "apikey secret-key" {
			t.Errorf("Authorization = %q", got)
		}
		if request.URL.Query().Get("apikey") != "" {
			t.Error("API key leaked into query")
		}
		switch request.URL.Path {
		case "/forex_pairs":
			return jsonResponse(`{"count":1,"data":[{"symbol":"USD/CNY"}],"status":"ok"}`), nil
		case "/quote":
			if got := request.URL.Query().Get("symbol"); got != "USD/CNY" {
				t.Errorf("symbol = %q", got)
			}
			return jsonResponse(`{"symbol":"USD/CNY","timestamp":1786534320,"close":"7.18500","is_market_open":true}`), nil
		default:
			t.Fatalf("unexpected path: %s", request.URL.Path)
			return nil, nil
		}
	})}

	client := NewRESTClient("https://api.example.invalid", "secret-key", nil, 8, httpClient, nil)
	quote, err := client.FetchQuote(context.Background(), provider.Subscription{CategoryCode: "forex", Market: "GB", Symbol: "USDCNY"})
	if err != nil {
		t.Fatal(err)
	}
	if quote.LastPriceText != "7.18500" || quote.Ts != 1786534320000 || quote.Authority != "twelvedata-rest" {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestRESTFetchQuoteRejectsMismatchedSymbol(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/forex_pairs" {
			return jsonResponse(`{"count":1,"data":[{"symbol":"USD/CNY"}],"status":"ok"}`), nil
		}
		return jsonResponse(`{"symbol":"EUR/USD","timestamp":1786534320,"close":"1.10","is_market_open":true}`), nil
	})}
	client := NewRESTClient("https://api.example.invalid", "secret-key", nil, 8, httpClient, nil)
	_, err := client.FetchQuote(context.Background(), provider.Subscription{CategoryCode: "forex", Symbol: "USDCNY"})
	if err == nil {
		t.Fatal("mismatched response symbol was accepted")
	}
}

func TestRESTFetchQuoteRejectsSymbolMissingFromForexCatalog(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/forex_pairs" {
			t.Fatalf("unsupported symbol must not reach quote endpoint: %s", request.URL.Path)
		}
		return jsonResponse(`{"count":1,"data":[{"symbol":"USD/CNY"}],"status":"ok"}`), nil
	})}
	client := NewRESTClient("https://api.example.invalid", "secret-key", nil, 8, httpClient, nil)
	_, err := client.FetchQuote(context.Background(), provider.Subscription{CategoryCode: "forex", Symbol: "COFFEE"})
	if err == nil {
		t.Fatal("symbol missing from Twelve Data forex catalog was accepted")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
