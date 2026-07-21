package types

import (
	"encoding/json"
	"testing"
)

func TestUpstreamDataPreservesExactLastPrice(t *testing.T) {
	var data UpstreamData
	if err := json.Unmarshal([]byte(`{"ld":123.456789012345678901,"t":100}`), &data); err != nil {
		t.Fatal(err)
	}
	if data.LDText != "123.456789012345678901" {
		t.Fatalf("exact price=%q", data.LDText)
	}
	if data.LD == 0 || data.T != 100 {
		t.Fatalf("numeric fields were not decoded: %+v", data)
	}
}
