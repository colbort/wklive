package models

import "context"

type UserIdentityModel interface {
	tUserIdentityModel
	FindPage(ctx context.Context, tenantId int64, page int64, pageSize int64) ([]*TUserIdentity, error)
}
