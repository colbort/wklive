package tasks

import (
	"context"

	"wklive/common/tasks"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("market.SyncProducts", "同步Market产品", syncMarketProducts)
	cronx.Register("market.SyncKlines", "同步Market K线", syncMarketKlines)
}

func syncMarketProducts(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceMarket, tasks.ActionMarketSyncProducts)
}

func syncMarketKlines(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceMarket, tasks.ActionMarketSyncKlines)
}
