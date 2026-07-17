package logic

import "testing"

func TestParseFingerprintVisitorID(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "official fingerprint", raw: `{"visitorId":"visitor-1","version":"5.2.0","confidence":0.7}`, want: "visitor-1", ok: true},
		{name: "legacy fingerprint", raw: `{"platform":"MacIntel","osName":"macos"}`, ok: false},
		{name: "empty visitor id", raw: `{"visitorId":""}`, ok: false},
		{name: "invalid json", raw: `{`, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseFingerprintVisitorID(tt.raw)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("parseFingerprintVisitorID() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestBuildFingerprintMatchKey(t *testing.T) {
	first := buildFingerprintMatchKey("visitor-1")
	second := buildFingerprintMatchKey("visitor-1")
	different := buildFingerprintMatchKey("visitor-2")

	if first == "" {
		t.Fatal("expected a match key")
	}
	if first != second {
		t.Fatal("the same visitorId must produce the same match key")
	}
	if first == different {
		t.Fatal("different visitorIds must produce different match keys")
	}
}
