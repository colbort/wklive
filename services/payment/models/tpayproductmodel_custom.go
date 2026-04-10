package models

import "context"

type PayProductModel interface {
	tPayProductModel
	FindPage(ctx context.Context, platformId int64, cursor int64, limit int64) ([]*TPayProduct, int64, error)
}
