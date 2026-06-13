package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
)

func adminTenantWriteScopeResp(ctx context.Context, currentTenantId int64, notAllowedCode int32) (*common.RespBase, error) {
	_, base, err := adminTenantWriteScope(ctx, currentTenantId, notAllowedCode)
	return base, err
}

func adminTenantWriteScope(ctx context.Context, currentTenantId int64, notAllowedCode int32) (bool, *common.RespBase, error) {
	allowTenantUpdate, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(ctx, currentTenantId)
	if err != nil {
		return false, nil, i18n.StatusError(ctx, i18n.UserNotFound)
	}
	if forbidden {
		return false, helper.GetErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)), nil
	}
	if !allowed {
		return false, helper.GetErrResp(notAllowedCode, i18n.Translate(notAllowedCode, ctx)), nil
	}
	return allowTenantUpdate, nil, nil
}

func systemAdminWriteScopeResp(ctx context.Context) (*common.RespBase, error) {
	userType, err := utils.GetUserTypeFromMd(ctx)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.UserNotFound)
	}
	if userType != utils.SysUserTypeSystemAdmin {
		return helper.GetErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)), nil
	}
	return nil, nil
}
