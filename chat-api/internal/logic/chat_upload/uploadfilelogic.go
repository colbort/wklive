package chat_upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
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
	chatUploadDir     = "chat_uploads"
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
	return saveUploadFile(file, header)
}

func (l *UploadFileLogic) UploadedFilePath(rawURL string) (string, error) {
	return uploadedFilePath(rawURL)
}

func saveUploadFile(file multipart.File, header *multipart.FileHeader) (*types.UploadFileResp, error) {
	if file == nil || header == nil {
		return nil, fmt.Errorf("file is required")
	}
	if header.Size <= 0 || header.Size > MaxChatUploadSize {
		return nil, fmt.Errorf("file size exceeds limit")
	}
	if err := os.MkdirAll(chatUploadDir, 0o755); err != nil {
		return nil, err
	}

	originalName := filepath.Base(header.Filename)
	ext := strings.ToLower(filepath.Ext(originalName))
	filename := fmt.Sprintf("chat_%d%s", time.Now().UnixNano(), ext)
	targetPath := filepath.Join(chatUploadDir, filename)
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
			Msg: "",
		},
		Data: types.UploadFileData{
			Url:      "/" + chatUploadDir + "/" + filename,
			FileName: originalName,
			FileSize: header.Size,
			MimeType: firstNonEmpty(header.Header.Get("Content-Type"), "application/octet-stream"),
		},
	}, nil
}

func uploadedFilePath(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("file url is required")
	}
	if parsed, err := url.Parse(rawURL); err == nil && parsed.Path != "" {
		rawURL = parsed.Path
	}

	cleanPath := filepath.Clean("/" + strings.TrimLeft(rawURL, "/"))
	prefix := "/" + chatUploadDir + "/"
	if !strings.HasPrefix(cleanPath, prefix) {
		return "", fmt.Errorf("invalid file url")
	}

	filename := filepath.Base(cleanPath)
	if filename == "." || filename == string(filepath.Separator) {
		return "", fmt.Errorf("invalid file url")
	}
	targetPath := filepath.Join(chatUploadDir, filename)
	info, err := os.Stat(targetPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("invalid file url")
	}
	return targetPath, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
