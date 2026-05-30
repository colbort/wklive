package tasks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"wklive/proto/itick"
	"wklive/proto/system"
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
	itickConfig, err := loadItickConfig(ctx)
	if err != nil {
		return err
	}

	rpcCtx, cancel := context.WithTimeout(ctx, syncItickProductsRPCTimeout)
	defer cancel()

	result, err := global.ItickTaskCli.SyncProducts(rpcCtx, &itick.SyncProductsReq{
		ApiUrl:   itickConfig.ApiUrl,
		ApiToken: itickConfig.ApiToken,
		WsUrl:    itickConfig.WsUrl,
	})
	if err != nil {
		return err
	}
	if result.Base.Code != 200 {
		return fmt.Errorf("sync products failed, code: %d, message: %s", result.Base.Code, result.Base.Msg)
	}
	return nil
}

func syncItickKlines(ctx context.Context, job *models.SysJob) error {
	itickConfig, err := loadItickConfig(ctx)
	if err != nil {
		return err
	}

	rpcCtx, cancel := context.WithTimeout(ctx, syncItickKlinesRPCTimeout)
	defer cancel()

	result, err := global.ItickTaskCli.SyncKlines(rpcCtx, &itick.SyncKlinesReq{
		ApiUrl:   itickConfig.ApiUrl,
		ApiToken: itickConfig.ApiToken,
		WsUrl:    itickConfig.WsUrl,
	})
	if err != nil {
		return err
	}
	if result.Base.Code != 200 {
		return fmt.Errorf("sync klines failed, code: %d, message: %s", result.Base.Code, result.Base.Msg)
	}
	return nil
}

func loadItickConfig(ctx context.Context) (*system.ItickConfig, error) {
	config, err := global.ConfigModel.FindOneByTenantIdConfigKey(ctx, 0, sql.NullString{
		String: system.SysConfigType_ITICK_CONFIG.String(),
		Valid:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find itick config: %w", err)
	}

	itickConfig := new(system.ItickConfig)
	err = json.Unmarshal([]byte(config.ConfigValue.String), itickConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse itick config: %w", err)
	}
	return itickConfig, nil
}
