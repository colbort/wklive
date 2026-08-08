package adminlogic

import (
	"context"
	"testing"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"
)

func TestCreateOpLogRejectsUntrustedActor(t *testing.T) {
	logic := NewCreateOpLogLogic(context.Background(), &svc.ServiceContext{})
	_, err := logic.CreateOpLog(&system.CreateOpLogReq{
		UserId:   99,
		Username: "forged",
		Path:     "/admin/system/users/status",
	})
	if err == nil {
		t.Fatal("expected untrusted actor metadata to be rejected")
	}
}

func TestTruncateOpLogValueUsesRuneLimit(t *testing.T) {
	if got := truncateOpLogValue("中文abc", 3); got != "中文a" {
		t.Fatalf("truncateOpLogValue=%q", got)
	}
}
