package models

import "context"

type TenantPayChannelModel interface {
	tTenantPayChannelModel
	FindPage(ctx context.Context, tenantId int64, platformId int64, productId int64, accountId int64, keyword string, status int64, cursor int64, limit int64) ([]*TTenantPayChannel, int64, error)
}
