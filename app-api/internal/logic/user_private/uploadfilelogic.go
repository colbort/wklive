// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"strings"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/i18n"
	"wklive/common/storage"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFileLogic) UploadFile(file multipart.File, header *multipart.FileHeader) (resp *types.UploadFileResp, err error) {
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		return &types.UploadFileResp{
			RespBase: types.RespBase{
				Code: 400,
				Msg:  i18n.Translate(i18n.ParamError, l.ctx),
			},
		}, nil
	}

	tenantId := int64(0)
	key := system.SysConfigType_OBJECT_STORAGE
	cd, err := l.svcCtx.SystemCli.SysConfigDetail(l.ctx, &system.SysConfigDetailReq{
		TenantId:  &tenantId,
		ConfigKey: &key,
	})
	if err != nil {
		return logicutil.SystemErrorResp[types.UploadFileResp](l.ctx, err)
	}

	var config system.ObjectStorageConfig
	if err := json.Unmarshal([]byte(cd.Data.ConfigValue), &config); err != nil {
		return logicutil.SystemErrorResp[types.UploadFileResp](l.ctx, err)
	}

	storageCfg := toStorageConfig(&config)
	url, err := storage.UploadFile(l.ctx, file, header, storageCfg)
	if err != nil {
		return logicutil.SystemErrorResp[types.UploadFileResp](l.ctx, err)
	}

	return &types.UploadFileResp{
		RespBase: types.RespBase{Code: i18n.OK, Msg: i18n.Translate(i18n.OK, l.ctx)},
		Data:     types.UploadFileData{Url: url},
	}, nil
}

func toStorageConfig(config *system.ObjectStorageConfig) storage.Config {
	if config == nil {
		return storage.Config{}
	}

	return storage.Config{
		AliyunOss:  toAliyunOssConfig(config.GetAliyunOss()),
		TencentCos: toTencentCosConfig(config.GetTencentCos()),
		Minio:      toMinioConfig(config.GetMinio()),
		OssType:    storage.OssType(config.GetOssType()),
		OssDomain:  config.GetOssDomain(),
	}
}

func toAliyunOssConfig(config *system.AliyunOssConfig) *storage.AliyunOssConfig {
	if config == nil {
		return nil
	}

	return &storage.AliyunOssConfig{
		Endpoint:        config.GetEndpoint(),
		AccessKeyId:     config.GetAccessKeyId(),
		AccessKeySecret: config.GetAccessKeySecret(),
		BucketName:      config.GetBucketName(),
		BucketUrl:       config.GetBucketUrl(),
	}
}

func toTencentCosConfig(config *system.TencentCosConfig) *storage.TencentCosConfig {
	if config == nil {
		return nil
	}

	return &storage.TencentCosConfig{
		Market:     config.GetMarket(),
		SecretId:   config.GetSecretId(),
		SecretKey:  config.GetSecretKey(),
		BucketName: config.GetBucketName(),
		BucketUrl:  config.GetBucketUrl(),
	}
}

func toMinioConfig(config *system.MinioConfig) *storage.MinioConfig {
	if config == nil {
		return nil
	}

	return &storage.MinioConfig{
		Endpoint:        config.GetEndpoint(),
		AccessKeyId:     config.GetAccessKeyId(),
		AccessKeySecret: config.GetAccessKeySecret(),
		BucketName:      config.GetBucketName(),
		BucketUrl:       config.GetBucketUrl(),
	}
}
