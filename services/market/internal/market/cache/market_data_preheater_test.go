package cache

import (
	"testing"

	"wklive/services/market/internal/market/types"
)

func TestRestQuoteIsAuthoritativeAndExact(t *testing.T) {
	payload, ok := restPayload(types.TopicQuote, types.UpstreamData{LD: 1.25, LDText: "1.250000000000000001", T: 100}).(*types.QuotePayload)
	if !ok {
		t.Fatal("REST quote payload type mismatch")
	}
	if payload.Authority != "itick-rest" || payload.LastPriceText != "1.250000000000000001" {
		t.Fatalf("REST quote lost authority or precision: %+v", payload)
	}
}
