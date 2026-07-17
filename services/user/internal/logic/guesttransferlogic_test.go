package logic

import (
	"strings"
	"testing"
)

func TestNormalizeTransferOrigin(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "https origin", raw: "https://NEW.example.com/", want: "https://new.example.com", ok: true},
		{name: "localhost", raw: "http://localhost:5173", want: "http://localhost:5173", ok: true},
		{name: "private ip", raw: "http://192.168.10.167:5174", want: "http://192.168.10.167:5174", ok: true},
		{name: "insecure remote", raw: "http://new.example.com", ok: false},
		{name: "path rejected", raw: "https://new.example.com/path", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeTransferOrigin(tt.raw)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("normalizeTransferOrigin() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestRandomGuestTransferCode(t *testing.T) {
	first, err := randomGuestTransferCode()
	if err != nil {
		t.Fatal(err)
	}
	second, err := randomGuestTransferCode()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) < 40 {
		t.Fatalf("expected distinct high-entropy codes, got %q and %q", first, second)
	}
	if strings.ContainsAny(first, "+/=") {
		t.Fatalf("code must be URL-safe: %q", first)
	}
	if guestTransferRedisKey(first) == guestTransferRedisKey(second) {
		t.Fatal("different codes must have different Redis keys")
	}
}
