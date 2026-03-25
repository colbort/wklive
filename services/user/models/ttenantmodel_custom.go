package models

import "context"

type TenantModel interface {
	tTenantModel
	FindPage(ctx context.Context, tenantId int64, page int64, pageSize int64) ([]*TTenant, error)
}
