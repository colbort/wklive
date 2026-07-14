package calendar

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestHolidaySyncFetch(t *testing.T) {
	svc := NewHolidaySyncService(context.Background(), "https://api.itick.test", "token", nil, nil, nil, nil, 0)
	svc.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/symbol/v2/holidays" || r.URL.Query().Get("code") != "HK" {
			t.Fatalf("unexpected request: %s", r.URL.String())
		}
		if r.Header.Get("token") != "token" {
			t.Fatal("missing token")
		}
		body := `{"code":0,"msg":"ok","data":[{"c":"HK","r":"Hong Kong","d":"2026-01-01","t":"09:30-12:00|13:00-16:00","z":"Asia/Hong_Kong","v":"New Year's Day"}]}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	items, err := svc.fetch(context.Background(), "HK")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Date != "2026-01-01" || items[0].Timezone != "Asia/Hong_Kong" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestNormalizeHolidayTimezone(t *testing.T) {
	if got := normalizeHolidayTimezone("Asia/Mumbai"); got != "Asia/Kolkata" {
		t.Fatalf("unexpected timezone: %s", got)
	}
	if got := normalizeHolidayTimezone("Asia/Hong_Kong"); got != "Asia/Hong_Kong" {
		t.Fatalf("canonical timezone changed: %s", got)
	}
}
