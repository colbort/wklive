package tradermade

import (
	"testing"
	"time"
)

func TestDecodeRESTResponsePreservesExactMidpoint(t *testing.T) {
	response, err := decodeRESTResponse([]byte(`{
		"endpoint":"live",
		"quotes":[{"ask":7.18501,"base_currency":"USD","bid":7.18499,"mid":7.18500,"quote_currency":"CNY"}],
		"timestamp":1786534335
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := response.Quotes[0].symbol(); got != "USDCNY" {
		t.Fatalf("symbol = %q", got)
	}
	text, value, err := response.Quotes[0].midpoint()
	if err != nil {
		t.Fatal(err)
	}
	if text != "7.18500" || value != 7.185 {
		t.Fatalf("midpoint = %q %v", text, value)
	}
}

func TestWebsocketMidpointFallsBackToBidAsk(t *testing.T) {
	message := wsEnvelope{Bid: "7.1849", Ask: "7.1851"}
	text, value, err := message.midpoint()
	if err != nil {
		t.Fatal(err)
	}
	if text != "7.185" || value != 7.185 {
		t.Fatalf("midpoint = %q %v", text, value)
	}
}

func TestParseWSTimestampUsesUTC(t *testing.T) {
	got, err := parseWSTimestamp("20260812-11:32:01.035")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 8, 12, 11, 32, 1, 35*int(time.Millisecond), time.UTC).UnixMilli()
	if got != want {
		t.Fatalf("timestamp = %d, want %d", got, want)
	}
}
