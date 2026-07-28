package models

import "testing"

func TestApplyBizTradeEventDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   int64
		want int64
	}{
		{name: "zero uses current version", in: 0, want: 1},
		{name: "negative uses current version", in: -1, want: 1},
		{name: "explicit version is preserved", in: 2, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &TBizTradeEvent{PayloadVersion: tt.in}
			applyBizTradeEventDefaults(event)
			if event.PayloadVersion != tt.want {
				t.Fatalf("PayloadVersion = %d, want %d", event.PayloadVersion, tt.want)
			}
		})
	}
}

func TestApplyBizTradeEventDefaultsNil(t *testing.T) {
	applyBizTradeEventDefaults(nil)
}
