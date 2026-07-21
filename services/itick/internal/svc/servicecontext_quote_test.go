package svc

import (
	"testing"

	"wklive/services/itick/internal/market/types"
)

func TestValidateAuthoritativeQuoteInput(t *testing.T) {
	if err := validateAuthoritativeQuoteInput(&types.QuotePayload{Authority: "itick-ws", LastPriceText: "123.45"}); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []*types.QuotePayload{
		nil,
		{LastPriceText: "123.45"},
		{Authority: "itick-ws"},
	} {
		if err := validateAuthoritativeQuoteInput(payload); err == nil {
			t.Fatalf("expected invalid quote to be rejected: %+v", payload)
		}
	}
}
