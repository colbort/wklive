package realtime

import (
	"testing"
	"time"
)

func TestUserEventHubRoutesOnlyToMatchingTenantAndUser(t *testing.T) {
	hub := NewUserEventHub()
	matching, cancelMatching := hub.Subscribe(1, 12)
	defer cancelMatching()
	otherUser, cancelOtherUser := hub.Subscribe(1, 13)
	defer cancelOtherUser()
	otherTenant, cancelOtherTenant := hub.Subscribe(2, 12)
	defer cancelOtherTenant()

	want := UserEvent{TenantID: 1, UserID: 12, ID: "event-1"}
	hub.Publish(want)

	select {
	case got := <-matching:
		if got.ID != want.ID {
			t.Fatalf("unexpected event: %#v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("matching subscriber did not receive event")
	}
	assertNoUserEvent(t, otherUser)
	assertNoUserEvent(t, otherTenant)
}

func assertNoUserEvent(t *testing.T, events <-chan UserEvent) {
	t.Helper()
	select {
	case event := <-events:
		t.Fatalf("unexpected private order event: %#v", event)
	case <-time.After(20 * time.Millisecond):
	}
}
