package models

import "context"

type UserModel interface {
	tUserModel
	FindPage(ctx context.Context, tenantId int64, page int64, pageSize int64) ([]*TUser, error)
}
