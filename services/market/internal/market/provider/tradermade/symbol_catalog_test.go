package tradermade

import (
	"context"
	"net/http"
	"testing"
)

func TestRESTSymbolCatalogUsesCurrencyCodesAndCFDDirectory(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/api/v1/live_currencies_list":
			return jsonHTTPResponse(`{"available_currencies":{"USD":"US Dollar","CNY":"Chinese Yuan"}}`), nil
		case "/api/v1/cfd_list":
			return jsonHTTPResponse(`{"available_cfds":{"AAPL":"Apple"}}`), nil
		default:
			t.Fatalf("unexpected path %s", request.URL.Path)
			return nil, nil
		}
	})}
	catalog := newRESTSymbolCatalog("https://marketdata.example/api/v1", "key", httpClient)
	if err := catalog.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	for input, expected := range map[string]string{"USD/CNY": "USDCNY", "CNYUSD": "CNYUSD", "aapl": "AAPL"} {
		actual, err := catalog.Resolve(input)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", input, err)
		}
		if actual != expected {
			t.Fatalf("Resolve(%q) = %q, want %q", input, actual, expected)
		}
	}
	if _, err := catalog.Resolve("EURUSD"); err == nil {
		t.Fatal("unlisted currency should not resolve")
	}
}

func TestStreamSymbolCatalogUsesExactListAndVerifiedCFDAlias(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/api/v1/streaming_currencies_list":
			return jsonHTTPResponse(`{"available_currencies":["USDCNY","AAPLUSD","XAUUSD"]}`), nil
		case "/api/v1/cfd_list":
			return jsonHTTPResponse(`{"available_cfds":{"AAPL":"Apple","COFFEE":"Coffee"}}`), nil
		default:
			t.Fatalf("unexpected path %s", request.URL.Path)
			return nil, nil
		}
	})}
	catalog := newStreamSymbolCatalog("https://marketdata.example/api/v1", "key", httpClient)
	if err := catalog.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if actual, err := catalog.Resolve("AAPL"); err != nil || actual != "AAPLUSD" {
		t.Fatalf("AAPL alias = %q, %v", actual, err)
	}
	if actual, err := catalog.Internal("AAPLUSD"); err != nil || actual != "AAPL" {
		t.Fatalf("AAPLUSD reverse alias = %q, %v", actual, err)
	}
	if actual, err := catalog.Resolve("USD/CNY"); err != nil || actual != "USDCNY" {
		t.Fatalf("USDCNY = %q, %v", actual, err)
	}
	if _, err := catalog.Resolve("COFFEE"); err == nil {
		t.Fatal("CFD missing from streaming directory must not be guessed")
	}
}
