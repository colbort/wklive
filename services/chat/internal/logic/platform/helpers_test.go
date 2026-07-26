package platformlogic

import (
	"context"
	"strconv"
	"testing"

	"wklive/common/i18n"
	"wklive/common/utils"

	"google.golang.org/grpc/metadata"
)

func TestPlatformScope(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		userType  int64
		wantAllow bool
	}{
		{name: "system administrator", domain: utils.SubjectDomainSystem, userType: utils.SysUserTypeSystemAdmin, wantAllow: true},
		{name: "chat merchant with colliding numeric type", domain: utils.SubjectDomainChat, userType: utils.SysUserTypeSystemAdmin},
		{name: "system tenant owner", domain: utils.SubjectDomainSystem, userType: utils.SysUserTypeTenantOwner},
		{name: "missing domain", userType: utils.SysUserTypeSystemAdmin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pairs := []string{utils.CtxKeyUserType, strconv.FormatInt(tt.userType, 10)}
			if tt.domain != "" {
				pairs = append(pairs, utils.CtxKeySubjectDomain, tt.domain)
			}
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(pairs...))
			base, err := platformScope(ctx)
			if err != nil {
				t.Fatalf("platformScope() error = %v", err)
			}
			if tt.wantAllow {
				if base != nil {
					t.Fatalf("platformScope() denied allowed caller: %+v", base)
				}
				return
			}
			if base == nil || base.Code != i18n.PermissionDenied {
				t.Fatalf("platformScope() base = %+v, want permission denied", base)
			}
		})
	}
}
