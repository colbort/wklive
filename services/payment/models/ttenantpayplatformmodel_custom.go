package models

import "context"

type TenantPayPlatformModel interface {
	tTenantPayPlatformModel
	FindPage(ctx context.Context, tenantId int64, platformId int64, status int64, openStatus int64, cursor int64, limit int64) ([]*TTenantPayPlatform, int64, error)
}
