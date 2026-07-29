package svc

import (
	"testing"

	"wklive/services/market/internal/market/types"
)

func TestValidateAuthoritativeQuoteInput(t *testing.T) {
	if err := validateAuthoritativeQuoteInput(&types.QuotePayload{Authority: "market-ws", LastPriceText: "123.45"}); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []*types.QuotePayload{
		nil,
		{LastPriceText: "123.45"},
		{Authority: "market-ws"},
	} {
		if err := validateAuthoritativeQuoteInput(payload); err == nil {
			t.Fatalf("expected invalid quote to be rejected: %+v", payload)
		}
	}
}
