package tasklogic

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestTruncateReconciliationDetailPreservesUTF8(t *testing.T) {
	input := strings.Repeat("差", 1001)
	got := truncateReconciliationDetail(input)
	if !utf8.ValidString(got) {
		t.Fatal("truncated detail is not valid UTF-8")
	}
	if len([]rune(got)) != 1000 {
		t.Fatalf("rune length=%d want=1000", len([]rune(got)))
	}
	if got := truncateReconciliationDetail("  ok  "); got != "ok" {
		t.Fatalf("trimmed detail=%q want ok", got)
	}
}

func TestReconciliationBusinessDateUsesPreviousUTCDay(t *testing.T) {
	now := time.Date(2026, 8, 1, 8, 5, 15, 0, time.FixedZone("UTC+8", 8*60*60))
	if got := reconciliationBusinessDate(now); got != "2026-07-31" {
		t.Fatalf("business date=%s want 2026-07-31", got)
	}
}
