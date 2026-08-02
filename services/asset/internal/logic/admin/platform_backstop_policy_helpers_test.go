package adminlogic

import (
	"testing"

	"wklive/proto/asset"
)

func validBackstopPolicyDraft(mode asset.PlatformBackstopMode) *asset.CreatePlatformBackstopPolicyReq {
	in := &asset.CreatePlatformBackstopPolicyReq{
		TenantId: 900101, Coin: " usdt ", RequestNo: "BST-POLICY-1", Mode: mode,
		PerRequestLimit: "10", DailyLimit: "100", BalanceFloor: "0",
		EffectiveFrom: 2000, EffectiveUntil: 3000,
		Reason: "bounded option liquidation backstop", EvidenceRef: "test://bst-policy-1",
	}
	if mode == asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_CREDIT_FLOOR {
		in.BalanceFloor = "-50"
	}
	if mode == asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_DISABLED {
		in.PerRequestLimit = "0"
		in.DailyLimit = "0"
	}
	return in
}

func TestValidateBackstopPolicyDraftModesAndBounds(t *testing.T) {
	for _, mode := range []asset.PlatformBackstopMode{
		asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_DISABLED,
		asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_PREFUNDED,
		asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_CREDIT_FLOOR,
	} {
		in := validBackstopPolicyDraft(mode)
		coin, requestNo, reason, perRequest, daily, floor, err := validateBackstopPolicyDraft(in, 1000)
		if err != nil {
			t.Fatalf("mode %v rejected: %v", mode, err)
		}
		if coin != "USDT" || requestNo != in.RequestNo || reason != in.Reason {
			t.Fatalf("mode %v normalization mismatch: %q %q %q", mode, coin, requestNo, reason)
		}
		if mode == asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_DISABLED &&
			(!perRequest.IsZero() || !daily.IsZero() || !floor.IsZero()) {
			t.Fatalf("disabled limits were not normalized to zero")
		}
	}
	tests := []struct {
		name   string
		mutate func(*asset.CreatePlatformBackstopPolicyReq)
	}{
		{name: "unknown mode", mutate: func(in *asset.CreatePlatformBackstopPolicyReq) { in.Mode = 0 }},
		{name: "prefunded negative floor", mutate: func(in *asset.CreatePlatformBackstopPolicyReq) { in.BalanceFloor = "-1" }},
		{name: "per request above daily", mutate: func(in *asset.CreatePlatformBackstopPolicyReq) { in.PerRequestLimit = "101" }},
		{name: "already effective", mutate: func(in *asset.CreatePlatformBackstopPolicyReq) { in.EffectiveFrom = 1000 }},
		{name: "duration above maximum", mutate: func(in *asset.CreatePlatformBackstopPolicyReq) {
			in.EffectiveUntil = in.EffectiveFrom + maxBackstopPolicyDuration + 1
		}},
		{name: "fraction precision above decimal", mutate: func(in *asset.CreatePlatformBackstopPolicyReq) { in.PerRequestLimit = "0.0000000000000000001" }},
		{name: "integer precision above decimal", mutate: func(in *asset.CreatePlatformBackstopPolicyReq) { in.DailyLimit = "1000000000000000000" }},
		{name: "missing evidence", mutate: func(in *asset.CreatePlatformBackstopPolicyReq) { in.EvidenceRef = "" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := validBackstopPolicyDraft(asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_PREFUNDED)
			tc.mutate(in)
			if _, _, _, _, _, _, err := validateBackstopPolicyDraft(in, 1000); err == nil {
				t.Fatalf("invalid draft accepted: %+v", in)
			}
		})
	}
}
