package assetlogic

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFreezeAllowsUnfreeze(t *testing.T) {
	tests := []struct {
		name   string
		status int64
		remain string
		want   bool
	}{
		{name: "freezing", status: 1, remain: "1", want: true},
		{name: "partially released", status: 2, remain: "1", want: true},
		{name: "legacy deducted with remainder", status: 4, remain: "0.1", want: true},
		{name: "deducted without remainder", status: 4, remain: "0", want: false},
		{name: "unfrozen terminal", status: 3, remain: "1", want: false},
		{name: "closed terminal", status: 5, remain: "1", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := freezeAllowsUnfreeze(tt.status, decimal.RequireFromString(tt.remain)); got != tt.want {
				t.Fatalf("freezeAllowsUnfreeze(%d, %s)=%t want=%t", tt.status, tt.remain, got, tt.want)
			}
		})
	}
}
