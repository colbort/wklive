package helpers

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestValidateCrossMarginOpeningRisk(t *testing.T) {
	tests := []struct {
		name string
		args []decimal.Decimal
		want string
	}{
		{
			name: "healthy",
			args: decimals("100", "80", "20", "5", "-10", "10", "1", "1"),
		},
		{
			name: "negative pnl consumes capacity",
			args: decimals("100", "20", "20", "5", "-15", "10", "1", "1"),
			want: "insufficient cross margin",
		},
		{
			name: "post fee equity must cover maintenance",
			args: decimals("5", "5", "0", "4", "0", "1", "1", "1"),
			want: "risk limit exceeded",
		},
		{
			name: "initial margin must cover order maintenance",
			args: decimals("100", "100", "0", "0", "0", "1", "0", "2"),
			want: "below maintenance margin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCrossMarginOpeningRisk(
				tt.args[0], tt.args[1], tt.args[2], tt.args[3],
				tt.args[4], tt.args[5], tt.args[6], tt.args[7],
			)
			if tt.want == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
				t.Fatalf("error=%v, want substring %q", err, tt.want)
			}
		})
	}
}

func decimals(values ...string) []decimal.Decimal {
	result := make([]decimal.Decimal, 0, len(values))
	for _, value := range values {
		result = append(result, decimal.RequireFromString(value))
	}
	return result
}
