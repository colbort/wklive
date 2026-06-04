// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_public

import (
	"context"
	"encoding/json"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
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
	key := system.SysConfigType_SYSTEM_CORE
	cd, err := logicutil.Proxy[system.SysConfigDetailResp](l.ctx, &system.SysConfigDetailReq{
		TenantId:  &tenantId,
		ConfigKey: &key,
	}, l.svcCtx.SystemCli.SysConfigDetail)
	if err != nil {
		return logicutil.SystemErrorResp[types.GetSystemCoreResp](l.ctx, err)
	}
	if cd.GetBase().GetCode() != 200 || cd.GetData() == nil {
		return &types.GetSystemCoreResp{
			RespBase: types.RespBase{
				Code: cd.GetBase().GetCode(),
				Msg:  cd.GetBase().GetMsg(),
			},
		}, nil
	}

	var core system.SystemCore
	err = json.Unmarshal([]byte(cd.Data.ConfigValue), &core)
	if err != nil {
		return logicutil.SystemErrorResp[types.GetSystemCoreResp](l.ctx, err)
	}

	key = system.SysConfigType_OBJECT_STORAGE
	cd, err = logicutil.Proxy[system.SysConfigDetailResp](l.ctx, &system.SysConfigDetailReq{
		TenantId:  &tenantId,
		ConfigKey: &key,
	}, l.svcCtx.SystemCli.SysConfigDetail)
	if err != nil {
		return logicutil.SystemErrorResp[types.GetSystemCoreResp](l.ctx, err)
	}
	if cd.GetBase().GetCode() != 200 || cd.GetData() == nil {
		return &types.GetSystemCoreResp{
			RespBase: types.RespBase{
				Code: cd.GetBase().GetCode(),
				Msg:  cd.GetBase().GetMsg(),
			},
		}, nil
	}
	var storage system.ObjectStorageConfig
	err = json.Unmarshal([]byte(cd.Data.ConfigValue), &storage)
	if err != nil {
		return logicutil.SystemErrorResp[types.GetSystemCoreResp](l.ctx, err)
	}
	assetUrl := ""
	switch storage.OssType {
	case 1:
		assetUrl = storage.AliyunOss.BucketUrl
	case 2:
		assetUrl = storage.TencentCos.BucketUrl
	case 3:
		assetUrl = storage.Minio.BucketUrl
	}
	return &types.GetSystemCoreResp{
		RespBase: types.RespBase{
			Code: cd.Base.Code,
			Msg:  cd.Base.Msg,
		},
		Data: types.GetSystemCore{
			SiteName: core.SiteName,
			SiteLogo: core.SiteLogo,
			AssetUrl: assetUrl,
		},
	}, nil
}
