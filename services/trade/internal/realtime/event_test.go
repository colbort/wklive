package realtime

import "testing"

func TestPartitionKeyPrefersBusinessEntity(t *testing.T) {
	event := Event{EventNo: "event-1", BizID: "fill-1"}
	if got := partitionKey(event); got != "fill-1" {
		t.Fatalf("partitionKey() = %q, want fill-1", got)
	}
}

func TestPartitionKeyFallsBackToEventNumber(t *testing.T) {
	event := Event{EventNo: "event-1"}
	if got := partitionKey(event); got != "event-1" {
		t.Fatalf("partitionKey() = %q, want event-1", got)
	}
}
