package logicutil

import (
	"context"

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
