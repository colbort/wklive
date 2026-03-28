package tasks

import (
	"context"
	"fmt"
	"wklive/proto/itick"
	"wklive/services/system/internal/global"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("itick.SyncProducts", "同步Itick产品", syncItickProducts)
}

func syncItickProducts(ctx context.Context, job *models.SysJob) error {
	result, err := global.ItickTaskCli.SyncProducts(ctx, &itick.SyncProductsReq{})
	if err != nil {
		return err
	}
	if result.Base.Code != 0 {
		err = fmt.Errorf("sync products failed, code: %d, message: %s", result.Base.Code, result.Base.Msg)
	}
	return nil
}
