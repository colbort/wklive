package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
)

func adminTenantWriteScopeResp(ctx context.Context, currentTenantId int64, notAllowedCode int32) (*common.RespBase, error) {
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(ctx, currentTenantId)
	if err != nil {
		return nil, i18n.StatusError(ctx, i18n.UserNotFound)
	}
	if forbidden {
		return helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)), nil
	}
	if !allowed {
		return helper.ErrResp(notAllowedCode, i18n.Translate(notAllowedCode, ctx)), nil
	}
	return nil, nil
}
