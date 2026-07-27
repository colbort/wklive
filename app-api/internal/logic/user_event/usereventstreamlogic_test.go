package user_event

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"wklive/app-api/internal/types"
)

func TestTenantIDFromJwtContext(t *testing.T) {
	tests := []struct {
		name   string
		expand string
		want   int64
	}{
		{name: "legacy tenant id", expand: `{"tenantId":12}`, want: 12},
		{name: "short tenant id", expand: `{"tid":13}`, want: 13},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "expand", test.expand)
			got, err := tenantIDFromJwtContext(ctx)
			if err != nil {
				t.Fatalf("tenantIDFromJwtContext() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("tenantIDFromJwtContext() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestConnectedEventOmitsBusinessData(t *testing.T) {
	output, err := json.Marshal(&types.UserEventStreamResp{
		EventType: "connected",
		ServerTs:  1,
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(output), `"data"`) {
		t.Fatalf("connected event contains empty business data: %s", output)
	}
}
