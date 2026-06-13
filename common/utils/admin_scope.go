package utils

import (
	"context"
)

func ResolveAdminTenantWriteScopeFromMd(ctx context.Context, currentTenantId int64) (bool, bool, bool, error) {
	userType, err := GetUserTypeFromMd(ctx)
	if err != nil {
		return false, false, false, err
	}

	switch userType {
	case SysUserTypeSystemAdmin:
		return true, true, false, nil
	case SysUserTypeTenantOwner, SysUserTypeTenantAdmin:
		tenantId, err := GetTenantIdFromMd(ctx)
		if err != nil {
			return false, false, false, err
		}
		return false, currentTenantId == tenantId, false, nil
	default:
		return false, false, true, nil
	}
}
