package ws

import (
	"testing"

	"wklive/common/notify"
)

func TestConnectionCanReceiveUsesTenantAndPermission(t *testing.T) {
	connection := NewConnection(
		nil,
		nil,
		10,
		"tenant-admin",
		7,
		false,
		[]string{"market:snapshot-outbox:list"},
		nil,
	)

	if !connection.CanReceive(notify.Event{
		Type:     "snapshot_outbox",
		TenantID: 7,
	}) {
		t.Fatal("expected matching tenant and permission to receive event")
	}
	if connection.CanReceive(notify.Event{
		Type:     "snapshot_outbox",
		TenantID: 8,
	}) {
		t.Fatal("cross-tenant event must be rejected")
	}
	if connection.CanReceive(notify.Event{
		Type:     "contract_reconciliation",
		TenantID: 7,
	}) {
		t.Fatal("event without required permission must be rejected")
	}
	if connection.CanReceive(notify.Event{
		Type:     "unknown",
		TenantID: 7,
	}) {
		t.Fatal("unknown event type must be rejected for tenant admin")
	}
}

func TestConnectionCanReceiveAllowsSystemAdmin(t *testing.T) {
	connection := NewConnection(nil, nil, 1, "admin", 0, true, nil, nil)
	if !connection.CanReceive(notify.Event{
		Type:     "unknown",
		TenantID: 99,
	}) {
		t.Fatal("system admin must receive platform notifications")
	}
}

func TestConnectionCanReceiveRejectsGlobalOperationalEventForTenantAdmin(t *testing.T) {
	connection := NewConnection(
		nil,
		nil,
		10,
		"tenant-admin",
		7,
		false,
		[]string{"market:price-formula:list"},
		nil,
	)
	if connection.CanReceive(notify.Event{
		Type:     "price_engine_input",
		TenantID: 0,
	}) {
		t.Fatal("platform operational event must be restricted to system admin")
	}
}
