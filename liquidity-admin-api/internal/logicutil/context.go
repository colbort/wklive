package logicutil

import (
	"context"
	"fmt"

	"wklive/common/utils"
)

func Identity(ctx context.Context) (tenantID, userID int64, err error) {
	tenantID, err = utils.GetTenantIdFromCtx(ctx)
	if err != nil {
		return 0, 0, err
	}
	userID, err = utils.GetUserIdFromCtx(ctx)
	return tenantID, userID, err
}

func TenantID(ctx context.Context) (int64, error) {
	return utils.GetTenantIdFromCtx(ctx)
}

// ResolveTenantID allows a system-side liquidity administrator (tenant 0) to
// operate on the tenant explicitly selected in the request. Tenant users are
// always restricted to the tenant carried by their token.
func ResolveTenantID(ctx context.Context, requested int64) (int64, error) {
	tokenTenantID, err := TenantID(ctx)
	if err != nil {
		return 0, err
	}
	if tokenTenantID > 0 {
		if requested > 0 && requested != tokenTenantID {
			return 0, fmt.Errorf("tenant access denied")
		}
		return tokenTenantID, nil
	}
	if requested <= 0 {
		return 0, fmt.Errorf("tenant_id is required")
	}
	return requested, nil
}
