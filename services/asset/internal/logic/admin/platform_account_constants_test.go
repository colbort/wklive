package adminlogic

import "testing"

func TestIsConfigurablePlatformAccountType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		accountType string
		want        bool
	}{
		{name: "insurance fund", accountType: insuranceFundAccountType, want: true},
		{name: "funding difference", accountType: fundingDifferenceAccountType, want: true},
		{name: "fee revenue", accountType: feeRevenueAccountType, want: true},
		{name: "option backstop", accountType: optionBackstopAccountType, want: true},
		{name: "unknown", accountType: "UNKNOWN", want: false},
		{name: "empty", accountType: "", want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isConfigurablePlatformAccountType(tt.accountType); got != tt.want {
				t.Fatalf("isConfigurablePlatformAccountType(%q) = %v, want %v", tt.accountType, got, tt.want)
			}
		})
	}
}
