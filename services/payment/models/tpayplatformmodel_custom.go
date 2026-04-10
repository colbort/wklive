package models

import "context"

type PayPlatformModel interface {
	tPayPlatformModel
	FindPage(ctx context.Context, keyword string, platformCode string, platformType int64, status int64, cursor int64, limit int64) ([]*TPayPlatform, int64, error)
}
