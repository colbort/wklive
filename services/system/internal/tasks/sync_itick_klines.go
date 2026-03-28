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
	cronx.Register("itick.SyncKlines", "同步Itick K线", syncItickKlines)
}

func syncItickKlines(ctx context.Context, job *models.SysJob) error {
	result, err := global.ItickTaskCli.SyncKlines(ctx, &itick.SyncKlinesReq{})
	if err != nil {
		return err
	}
	if result.Base.Code != 0 {
		err = fmt.Errorf("sync klines failed, code: %d, message: %s", result.Base.Code, result.Base.Msg)
	}
	return nil
}
