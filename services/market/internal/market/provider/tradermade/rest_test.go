package tradermade

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"wklive/services/market/internal/market/provider"
)

func TestRESTClientFetchQuote(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/api/v1/live_currencies_list":
			return jsonHTTPResponse(`{"available_currencies":{"USD":"US Dollar","CNY":"Chinese Yuan"}}`), nil
		case "/api/v1/cfd_list":
			return jsonHTTPResponse(`{"available_cfds":{}}`), nil
		}
		if got := request.URL.Query().Get("currency"); got != "USDCNY" {
			t.Errorf("currency = %q", got)
		}
		if got := request.URL.Query().Get("api_key"); got != "rest-key" {
			t.Errorf("api_key = %q", got)
		}
		return jsonHTTPResponse(`{"quotes":[{"ask":7.2,"bid":7.1,"mid":7.15,"base_currency":"USD","quote_currency":"CNY"}],"timestamp":1786534335}`), nil
	})}

	client := NewRESTClient("https://marketdata.example/api/v1", "rest-key", httpClient, nil)
	quote, err := client.FetchQuote(context.Background(), provider.Subscription{CategoryCode: "forex", Market: "GB", Symbol: "USDCNY"})
	if err != nil {
		t.Fatal(err)
	}
	if quote.LastPriceText != "7.15" || quote.Ts != 1786534335000 || quote.Authority != "tradermade-rest" {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestRESTClientRejectsApplicationErrorWithHTTP200(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/api/v1/live_currencies_list" {
			return jsonHTTPResponse(`{"available_currencies":{"USD":"US Dollar","CNY":"Chinese Yuan"}}`), nil
		}
		if request.URL.Path == "/api/v1/cfd_list" {
			return jsonHTTPResponse(`{"available_cfds":{}}`), nil
		}
		return jsonHTTPResponse(`{"code":401,"message":"api key invalid"}`), nil
	})}
	client := NewRESTClient("https://marketdata.example/api/v1", "bad-key", httpClient, nil)
	_, err := client.FetchQuote(context.Background(), provider.Subscription{CategoryCode: "forex", Symbol: "USDCNY"})
	if err == nil {
		t.Fatal("expected application-level error")
	}
}

func TestRESTClientDoesNotLeakAPIKeyInTransportError(t *testing.T) {
	const secret = "super-secret-key"
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return nil, &url.Error{Op: "Get", URL: request.URL.String(), Err: errors.New("connection reset")}
	})}
	client := NewRESTClient("https://marketdata.example/api/v1", secret, httpClient, nil)
	_, err := client.FetchQuote(context.Background(), provider.Subscription{CategoryCode: "forex", Symbol: "USDCNY"})
	if err == nil {
		t.Fatal("expected transport error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("API key leaked in error: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func jsonHTTPResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
