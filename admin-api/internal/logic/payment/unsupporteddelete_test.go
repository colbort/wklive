package payment

import "testing"

func TestUnsupportedDeleteFailsClosed(t *testing.T) {
	resp := unsupportedDelete("test resource")
	if resp.Code == 0 || resp.Code == 200 {
		t.Fatalf("unsupported deletion must not report success: %+v", resp)
	}
	if resp.Msg == "" {
		t.Fatal("unsupported deletion must explain why it was rejected")
	}
}
