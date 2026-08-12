package helpers

import (
	"context"
	"errors"
	"testing"

	"wklive/common/i18n"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsDefinitiveAssetFreezeRejection(t *testing.T) {
	if !IsDefinitiveAssetFreezeRejection(i18n.StatusError(context.Background(), i18n.InsufficientAvailableBalance)) {
		t.Fatal("explicit insufficient balance rejection must be definitive")
	}
	for _, err := range []error{
		errors.New("transport failed"),
		status.Error(codes.Unavailable, "asset rpc unavailable"),
		status.Error(codes.DeadlineExceeded, "asset rpc timed out"),
	} {
		if IsDefinitiveAssetFreezeRejection(err) {
			t.Fatalf("unknown RPC outcome must not be definitive: %v", err)
		}
	}
}
