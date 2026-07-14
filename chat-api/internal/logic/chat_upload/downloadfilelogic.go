package chat_upload

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"chat-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDownloadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadFileLogic {
	return &DownloadFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DownloadFileLogic) DownloadedFilePath(rawURL string) (string, error) {
	return downloadedFilePath(l.dir(), rawURL)
}

func (l *DownloadFileLogic) dir() string {
	dir := strings.TrimSpace(l.svcCtx.Config.ChatUploadDir)
	if dir == "" {
		return chatUploadURLPath
	}
	return dir
}

func downloadedFilePath(uploadDir string, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("file url is required")
	}
	if parsed, err := url.Parse(rawURL); err == nil && parsed.Path != "" {
		rawURL = parsed.Path
	}

	cleanPath := filepath.Clean("/" + strings.TrimLeft(rawURL, "/"))
	prefix := "/" + chatUploadURLPath + "/"
	if !strings.HasPrefix(cleanPath, prefix) {
		return "", fmt.Errorf("invalid file url")
	}

	filename := filepath.Base(cleanPath)
	if filename == "." || filename == string(filepath.Separator) {
		return "", fmt.Errorf("invalid file url")
	}
	targetPath := filepath.Join(uploadDir, filename)
	info, err := os.Stat(targetPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("invalid file url")
	}
	return targetPath, nil
}
