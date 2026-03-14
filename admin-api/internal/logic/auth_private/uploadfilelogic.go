// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"context"
	"mime/multipart"
	"strings"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/storage"
	"wklive/common/utils"
	"wklive/proto/system"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/errorx"
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
				Msg:  "只允许上传图片文件",
			},
		}, nil
	}

	key := system.SysConfigType_OBJECT_STORAGE.String()
	cd, err := l.svcCtx.SystemCli.SysConfigDetail(l.ctx, &system.SysConfigDetailReq{
		ConfigKey: &key,
	})
	if err != nil {
		return nil, err
	}

	var config system.ObjectStorageConfig
	err = utils.StructToGoStruct(cd.Data.ConfigValue, &config)
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
		return nil, errorx.Wrap(err, "上传文件失败")
	}

	resp = &types.UploadFileResp{
		RespBase: types.RespBase{Code: 200, Msg: "上传成功"},
		Data: struct {
			Url string `json:"url"`
		}{Url: url},
	}
	return resp, nil
}
