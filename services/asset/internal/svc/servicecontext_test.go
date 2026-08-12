package svc

import (
	"context"
	"errors"
	"testing"

	"wklive/common/i18n"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAssetBusinessErrorAcceptable(t *testing.T) {
	if !assetBusinessErrorAcceptable(i18n.StatusError(context.Background(), i18n.InsufficientAvailableBalance)) {
		t.Fatal("insufficient balance must not poison the database breaker")
	}
	for _, code := range []codes.Code{
		codes.InvalidArgument,
		codes.FailedPrecondition,
		codes.AlreadyExists,
		codes.NotFound,
		codes.PermissionDenied,
	} {
		if !assetBusinessErrorAcceptable(status.Error(code, "business rejection")) {
			t.Fatalf("business code %s would poison the database breaker", code)
		}
	}
	for _, err := range []error{
		nil,
		errors.New("database failure"),
		status.Error(codes.Internal, "database failure"),
		status.Error(codes.Unavailable, "database unavailable"),
		status.Error(codes.DeadlineExceeded, "database timeout"),
	} {
		if assetBusinessErrorAcceptable(err) {
			t.Fatalf("system error accepted by database breaker: %v", err)
		}
	}
}
