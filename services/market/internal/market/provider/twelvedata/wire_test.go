package twelvedata

import "testing"

func TestDecodeRESTQuotePreservesExactPriceAndTimestamp(t *testing.T) {
	quote, err := decodeRESTQuote([]byte(`{
		"symbol":"USD/CNY",
		"timestamp":1786534320,
		"close":"7.18500",
		"is_market_open":true
	}`))
	if err != nil {
		t.Fatal(err)
	}
	text, value, err := quote.price()
	if err != nil {
		t.Fatal(err)
	}
	if text != "7.18500" || value != 7.185 {
		t.Fatalf("price = %q %v", text, value)
	}
	if quote.Timestamp != 1786534320 {
		t.Fatalf("timestamp = %d", quote.Timestamp)
	}
}

func TestDecodeRESTQuoteRejectsClosedMarket(t *testing.T) {
	_, err := decodeRESTQuote([]byte(`{
		"symbol":"USD/CNY",
		"timestamp":1786534320,
		"close":"7.18500",
		"is_market_open":false
	}`))
	if err == nil {
		t.Fatal("closed market quote was accepted")
	}
}

func TestRawDecimalAcceptsNumberAndString(t *testing.T) {
	for _, raw := range []string{`7.18500`, `"7.18500"`} {
		text, value, err := rawDecimal([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if text != "7.18500" || value != 7.185 {
			t.Fatalf("rawDecimal(%s) = %q %v", raw, text, value)
		}
	}
}
