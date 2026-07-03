package tasks

import (
	"context"

	"wklive/common/tasks"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("itick.SyncProducts", "同步Itick产品", syncItickProducts)
	cronx.Register("itick.SyncKlines", "同步Itick K线", syncItickKlines)
}

func syncItickProducts(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceItick, tasks.ActionItickSyncProducts)
}

func syncItickKlines(ctx context.Context, job *models.SysJob) error {
	return publishTask(ctx, job, tasks.ServiceItick, tasks.ActionItickSyncKlines)
}
