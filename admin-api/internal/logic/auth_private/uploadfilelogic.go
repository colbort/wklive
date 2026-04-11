// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/i18n"
	"wklive/common/storage"
	"wklive/proto/system"

	"github.com/jinzhu/copier"
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

/**
 * @description: 上传文件
 * @param {multipart.File} file 文件
 * @param {multipart.FileHeader} header 文件头
 * @return {*types.UploadFileResp} resp 响应
 * @return {*error} err 错误
 */
// 1、首先验证上传的文件是否为图片类型，如果不是图片类型，直接返回错误信息
// 2、获取管理后台配置的存储类型，进行相应的文件上传处理
// 3、如果上传成功，返回文件的访问URL；如果上传失败，返回错误信息
func (l *UploadFileLogic) UploadFile(file multipart.File, header *multipart.FileHeader) (resp *types.UploadFileResp, err error) {
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		return &types.UploadFileResp{
			RespBase: types.RespBase{
				Code: 400,
				Msg:  i18n.Translate(i18n.ParamError, l.ctx),
			},
		}, nil
	}

	key := system.SysConfigType_OBJECT_STORAGE
	cd, err := l.svcCtx.SystemCli.SysConfigDetail(l.ctx, &system.SysConfigDetailReq{
		ConfigKey: &key,
	})
	if err != nil {
		return nil, err
	}

	var config system.ObjectStorageConfig
	err = json.Unmarshal([]byte(cd.Data.ConfigValue), &config)
	if err != nil {
		return nil, err
	}

	// Convert proto config into the common storage config type.
	var storageCfg storage.Config
	err = copier.Copy(&storageCfg, &config)
	if err != nil {
		return nil, err
	}
	url, err := storage.UploadFile(l.ctx, file, header, storageCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Translate(i18n.InternalServerError, l.ctx), err)
	}

	resp = &types.UploadFileResp{
		RespBase: types.RespBase{Code: i18n.OK, Msg: i18n.Translate(i18n.OK, l.ctx)},
		Data: struct {
			Url string `json:"url"`
		}{Url: url},
	}
	return resp, nil
}
