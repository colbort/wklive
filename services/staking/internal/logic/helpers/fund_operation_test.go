package helpers

import (
	"testing"

	"wklive/proto/staking"
)

func TestShouldMarkEarlyRedeemed(t *testing.T) {
	now := int64(100)
	tests := []struct {
		name       string
		redeemType int64
		end        int64
		want       bool
	}{
		{name: "explicit early", redeemType: int64(staking.RedeemType_REDEEM_TYPE_EARLY), end: 50, want: true},
		{name: "manual before maturity", redeemType: int64(staking.RedeemType_REDEEM_TYPE_MANUAL), end: 101, want: true},
		{name: "manual flexible", redeemType: int64(staking.RedeemType_REDEEM_TYPE_MANUAL), end: 0, want: true},
		{name: "manual matured", redeemType: int64(staking.RedeemType_REDEEM_TYPE_MANUAL), end: 100, want: false},
		{name: "maturity", redeemType: int64(staking.RedeemType_REDEEM_TYPE_MATURITY), end: 100, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldMarkEarlyRedeemed(tt.redeemType, tt.end, now); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
