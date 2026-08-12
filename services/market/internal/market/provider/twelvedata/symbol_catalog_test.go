package twelvedata

import (
	"context"
	"net/http"
	"testing"
)

func TestSymbolCatalogMapsCanonicalToOriginalSymbol(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/forex_pairs" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		if request.URL.Query().Get("outputsize") != "5000" {
			t.Errorf("outputsize = %q", request.URL.Query().Get("outputsize"))
		}
		return jsonResponse(`{
			"count":2,
			"data":[{"symbol":"USD/CNY"},{"symbol":"XAU/USD"}],
			"status":"ok"
		}`), nil
	})}
	catalog := newSymbolCatalog("https://api.example.invalid", "secret-key", nil, httpClient)
	if err := catalog.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	for internal, want := range map[string]string{"USDCNY": "USD/CNY", "XAUUSD": "XAU/USD"} {
		got, err := catalog.Resolve(internal)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("Resolve(%q) = %q, want %q", internal, got, want)
		}
	}
	if _, err := catalog.Resolve("COFFEE"); err == nil {
		t.Fatal("catalog guessed an unsupported symbol")
	}
}
