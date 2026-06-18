package tasks

import (
	"context"
	"fmt"
	"time"

	"wklive/proto/itick"
	"wklive/services/system/internal/global"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

const (
	syncItickKlinesRPCTimeout   = 30 * time.Second
	syncItickProductsRPCTimeout = 30 * time.Minute
)

func init() {
	cronx.Register("itick.SyncProducts", "同步Itick产品", syncItickProducts)
	cronx.Register("itick.SyncKlines", "同步Itick K线", syncItickKlines)
}

func syncItickProducts(ctx context.Context, job *models.SysJob) error {
	rpcCtx, cancel := context.WithTimeout(ctx, syncItickProductsRPCTimeout)
	defer cancel()

	result, err := global.ItickTaskCli.SyncProducts(rpcCtx, &itick.SyncProductsReq{})
	if err != nil {
		return err
	}
	if result.Base.Code != 200 {
		return fmt.Errorf("sync products failed, code: %d, message: %s", result.Base.Code, result.Base.Msg)
	}
	return nil
}

func syncItickKlines(ctx context.Context, job *models.SysJob) error {
	rpcCtx, cancel := context.WithTimeout(ctx, syncItickKlinesRPCTimeout)
	defer cancel()

	result, err := global.ItickTaskCli.SyncKlines(rpcCtx, &itick.SyncKlinesReq{})
	if err != nil {
		return err
	}
	if result.Base.Code != 200 {
		return fmt.Errorf("sync klines failed, code: %d, message: %s", result.Base.Code, result.Base.Msg)
	}
	return nil
}
