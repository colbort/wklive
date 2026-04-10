package models

import "context"

type TenantPayAccountModel interface {
	tTenantPayAccountModel
	FindPage(ctx context.Context, tenantId int64, platformId int64, cursor int64, limit int64) ([]*TTenantPayAccount, int64, error)
}
