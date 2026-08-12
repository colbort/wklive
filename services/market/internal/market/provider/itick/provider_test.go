package itick

import (
	"context"
	"testing"

	"wklive/services/market/internal/market/provider"
)

func TestProviderIdentityAndValidation(t *testing.T) {
	var source provider.RealtimeProvider = New("", "", "", nil, nil, nil)
	if source.Code() != "itick" {
		t.Fatalf("unexpected provider code: %s", source.Code())
	}
	if source.Supports("forex") {
		t.Fatal("an unconfigured provider must not report category support")
	}
	if _, err := source.NewStream("forex"); err == nil {
		t.Fatal("an unconfigured provider must reject stream creation")
	}
	if _, err := source.FetchQuote(context.Background(), provider.Subscription{}); err == nil {
		t.Fatal("an unconfigured provider must reject quote fetches")
	}
}
