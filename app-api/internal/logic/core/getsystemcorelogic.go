// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package core

import (
	"context"
	"encoding/json"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/market"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type GetSystemCoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSystemCoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSystemCoreLogic {
	return &GetSystemCoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSystemCoreLogic) GetSystemCore() (resp *types.GetSystemCoreResp, err error) {
	tenantId := int64(0)
	coreKey := system.SysConfigType_SYSTEM_CORE
	storageKey := system.SysConfigType_OBJECT_STORAGE
	var coreConfig, storageConfig *system.SysConfigDetailResp
	var intervalsResult *market.KlineIntervalsResp
	err = mr.Finish(
		func() error {
			var callErr error
			coreConfig, callErr = l.svcCtx.SystemCli.SysConfigDetail(l.ctx, &system.SysConfigDetailReq{
				TenantId: &tenantId, ConfigKey: &coreKey,
			})
			return callErr
		},
		func() error {
			var callErr error
			storageConfig, callErr = l.svcCtx.SystemCli.SysConfigDetail(l.ctx, &system.SysConfigDetailReq{
				TenantId: &tenantId, ConfigKey: &storageKey,
			})
			return callErr
		},
		func() error {
			var callErr error
			intervalsResult, callErr = l.svcCtx.MarketCli.GetKlineIntervals(l.ctx, &market.Empty{})
			return callErr
		},
	)
	if err != nil {
		return logicutil.SystemErrorResp[types.GetSystemCoreResp](l.ctx, err)
	}
	var config system.SystemCore
	err = json.Unmarshal([]byte(coreConfig.GetData().GetConfigValue()), &config)
	if err != nil {
		return logicutil.SystemErrorResp[types.GetSystemCoreResp](l.ctx, err)
	}
	var storage system.ObjectStorageConfig
	err = json.Unmarshal([]byte(storageConfig.GetData().GetConfigValue()), &storage)
	if err != nil {
		return logicutil.SystemErrorResp[types.GetSystemCoreResp](l.ctx, err)
	}
	intervals := make([]types.Interval, 0)
	for _, item := range intervalsResult.GetData() {
		intervals = append(intervals, types.Interval{
			Name:  item.Name,
			KType: item.KType,
		})
	}
	data := types.SystemCore{
		IsCaptchaEnabled:  enableToBool(config.IsCaptchaEnabled),
		IsRegisterEnabled: enableToBool(config.IsRegisterEnabled),
		IsGuestEnabled:    enableToBool(config.IsGuestEnabled),
		IsCryptoEnabled:   enableToBool(config.IsCryptoEnabled),
		AssetUrl:          objectStorageAssetUrl(&storage),
		Intervals:         intervals,
		Options:           logicutil.CoreOptions(),
	}
	return &types.GetSystemCoreResp{
		RespBase: types.RespBase{
			Code: 200,
			Msg:  "获取系统配置成功",
		},
		Data: data,
	}, nil
}

func enableToBool(value common.Enable) bool {
	return value == common.Enable_ENABLE_ENABLED
}

func objectStorageAssetUrl(storage *system.ObjectStorageConfig) string {
	if storage == nil {
		return ""
	}
	switch storage.GetOssType() {
	case 1:
		if cfg := storage.GetAliyunOss(); cfg != nil {
			return cfg.GetBucketUrl()
		}
	case 2:
		if cfg := storage.GetTencentCos(); cfg != nil {
			return cfg.GetBucketUrl()
		}
	case 3:
		if cfg := storage.GetMinio(); cfg != nil {
			return cfg.GetBucketUrl()
		}
	}
	return storage.GetOssDomain()
}
