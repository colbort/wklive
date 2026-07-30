package ws

import (
	"net/http/httptest"
	"testing"
)

func TestOriginAllowed(t *testing.T) {
	config := Config{
		AllowedOrigins: []string{"https://admin.example.com"},
	}

	request := httptest.NewRequest("GET", "http://api.example.com/admin/ws/notifications", nil)
	request.Header.Set("Origin", "https://admin.example.com")
	if !OriginAllowed(request, config) {
		t.Fatal("configured origin must be allowed")
	}

	request.Header.Set("Origin", "https://other.example.com")
	if OriginAllowed(request, config) {
		t.Fatal("unconfigured cross-origin request must be rejected")
	}

	request.Header.Del("Origin")
	if OriginAllowed(request, config) {
		t.Fatal("missing origin must be rejected by default")
	}
	config.AllowMissingOrigin = true
	if !OriginAllowed(request, config) {
		t.Fatal("explicit missing-origin allowance must be honored")
	}
}

func TestOriginAllowedAcceptsSameHost(t *testing.T) {
	request := httptest.NewRequest("GET", "http://admin.example.com/admin/ws/notifications", nil)
	request.Host = "admin.example.com"
	request.Header.Set("Origin", "https://admin.example.com")
	if !OriginAllowed(request, Config{}) {
		t.Fatal("same host origin must be allowed")
	}
}
