package tasklogic

import "testing"

func TestMarkQuoteRetryGate(t *testing.T) {
	gate := newMarkQuoteRetryGate()
	const (
		key = "900101:990101:crypto:BA:BTCUSDT"
		now = int64(1_000_000)
	)

	if !gate.allow(key, now) {
		t.Fatal("new quote key must be attempted immediately")
	}

	gate.fail(key, now)
	if gate.allow(key, now) {
		t.Fatal("failed quote key must be held during backoff")
	}
	if gate.allow(key, now+markQuoteRetryBackoffMs-1) {
		t.Fatal("failed quote key must stay held until the boundary")
	}
	if !gate.allow(key, now+markQuoteRetryBackoffMs) {
		t.Fatal("failed quote key must retry at the boundary")
	}

	if !gate.allow(key+":other-source", now) {
		t.Fatal("one failed source must not suppress another source")
	}

	gate.success(key)
	if !gate.allow(key, now) {
		t.Fatal("a successful lookup must clear the backoff immediately")
	}
}
