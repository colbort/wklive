package chat_upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	MaxChatUploadSize = 100 << 20
	chatUploadURLPath = "chat_uploads"
)

type UploadFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UploadFileLogic) UploadFile(file multipart.File, header *multipart.FileHeader) (*types.UploadFileResp, error) {
	return saveUploadFile(l.dir(), file, header)
}

func (l *UploadFileLogic) dir() string {
	dir := strings.TrimSpace(l.svcCtx.Config.ChatUploadDir)
	if dir == "" {
		return chatUploadURLPath
	}
	return dir
}

func saveUploadFile(uploadDir string, file multipart.File, header *multipart.FileHeader) (*types.UploadFileResp, error) {
	if file == nil || header == nil {
		return nil, fmt.Errorf("file is required")
	}
	if header.Size <= 0 || header.Size > MaxChatUploadSize {
		return nil, fmt.Errorf("file size exceeds limit")
	}
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, err
	}

	originalName := filepath.Base(header.Filename)
	ext := strings.ToLower(filepath.Ext(originalName))
	filename := fmt.Sprintf("chat_%d%s", time.Now().UnixNano(), ext)
	targetPath := filepath.Join(uploadDir, filename)
	target, err := os.Create(targetPath)
	if err != nil {
		return nil, err
	}
	defer target.Close()

	if _, err := io.Copy(target, file); err != nil {
		return nil, err
	}

	return &types.UploadFileResp{
		RespBase: types.RespBase{
			Code: 200,
			Msg:  "",
		},
		Data: types.UploadFileData{
			Url:      "/" + chatUploadURLPath + "/" + filename,
			FileName: originalName,
			FileSize: header.Size,
			MimeType: firstNonEmpty(header.Header.Get("Content-Type"), "application/octet-stream"),
		},
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
