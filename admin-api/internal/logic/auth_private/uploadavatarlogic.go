// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadAvatarLogic) UploadAvatar(file multipart.File, header *multipart.FileHeader) (resp *types.UploadFileResp, err error) {
	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		return &types.UploadFileResp{
			RespBase: types.RespBase{
				Code: 400,
				Msg:  "只允许上传图片文件",
			},
		}, nil
	}

	// 读取文件内容并计算 md5
	data, err := io.ReadAll(file)
	if err != nil {
		return &types.UploadFileResp{
			RespBase: types.RespBase{
				Code: 500,
				Msg:  "读取文件失败",
			},
		}, nil
	}

	hash := md5.Sum(data)
	ext := filepath.Ext(header.Filename)
	fname := hex.EncodeToString(hash[:]) + ext

	// 检查文件夹是否存在，不存在则创建
	avatarDir := "./avatars"
	if _, err := os.Stat(avatarDir); os.IsNotExist(err) {
		err = os.MkdirAll(avatarDir, os.ModePerm)
		if err != nil {
			return &types.UploadFileResp{
				RespBase: types.RespBase{
					Code: 500,
					Msg:  "创建目录失败",
				},
			}, nil
		}
	}

	filePath := filepath.Join(avatarDir, fname)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，返回地址
		return &types.UploadFileResp{
			RespBase: types.RespBase{
				Code: 200,
				Msg:  "文件已存在",
			},
			Data: struct {
				Url string `json:"url"`
			}{
				Url: "/avatars/" + fname,
			},
		}, nil
	}

	// 保存文件
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return &types.UploadFileResp{
			RespBase: types.RespBase{
				Code: 500,
				Msg:  "文件保存失败",
			},
		}, nil
	}

	return &types.UploadFileResp{
		RespBase: types.RespBase{
			Code: 200,
			Msg:  "上传成功",
		},
		Data: struct {
			Url string `json:"url"`
		}{
			Url: "/avatars/" + fname,
		},
	}, nil
}
