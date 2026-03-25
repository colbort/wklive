package models

import "context"

type UserSecurityModel interface {
	tUserSecurityModel
	FindPage(ctx context.Context, tenantId int64, page int64, pageSize int64) ([]*TUserSecurity, error)
}
