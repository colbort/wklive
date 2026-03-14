// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	ext := filepath.Ext(header.Filename)
	fname := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	dst, err := os.Create("./avatars/" + fname)
	if err != nil {
		return &types.UploadFileResp{
			RespBase: types.RespBase{
				Code: 500,
				Msg:  "文件保存失败",
			},
		}, nil
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return &types.UploadFileResp{
			RespBase: types.RespBase{
				Code: 500,
				Msg:  "文件写入失败",
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
			Url: "./avatars/" + fname,
		},
	}, nil
}
