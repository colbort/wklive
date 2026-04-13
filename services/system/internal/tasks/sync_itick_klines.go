package tasks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"wklive/proto/itick"
	"wklive/proto/system"
	"wklive/services/system/internal/global"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

func init() {
	cronx.Register("itick.SyncKlines", "同步Itick K线", syncItickKlines)
}

func syncItickKlines(ctx context.Context, job *models.SysJob) error {
	config, err := global.ConfigModel.FindOneByConfigKey(ctx, sql.NullString{
		String: system.SysConfigType_ITICK_CONFIG.String(),
		Valid:  true,
	})
	var itickConfig system.ItickConfig
	err = json.Unmarshal([]byte(config.ConfigValue.String), &itickConfig)
	if err != nil {
		return fmt.Errorf("failed to find itick config: %w", err)
	}
	result, err := global.ItickTaskCli.SyncKlines(ctx, &itick.SyncKlinesReq{
		ApiUrl:   itickConfig.ApiUrl,
		ApiToken: itickConfig.ApiToken,
		WsUrl:    itickConfig.WsUrl,
	})
	if err != nil {
		return err
	}
	if result.Base.Code != 0 {
		err = fmt.Errorf("sync klines failed, code: %d, message: %s", result.Base.Code, result.Base.Msg)
	}
	return nil
}
