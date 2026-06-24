package chat_auth

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chat-admin-api/internal/logic/chat_auth"
	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

const maxAvatarUploadSize = 5 << 20

func UploadProfileAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxAvatarUploadSize)
		file, header, err := r.FormFile("avatar")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		ext, err := avatarFileExt(header.Filename)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		if err := os.MkdirAll("avatars", 0o755); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		filename := fmt.Sprintf("avatar_%d%s", time.Now().UnixNano(), ext)
		targetPath := filepath.Join("avatars", filename)
		target, err := os.Create(targetPath)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer target.Close()

		if _, err := io.Copy(target, file); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		resp, err := updateProfileAvatar(r, svcCtx, "/avatars/"+filename)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func avatarFileExt(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return ext, nil
	default:
		return "", fmt.Errorf("仅支持 jpg、png、webp 图片")
	}
}

func updateProfileAvatar(r *http.Request, svcCtx *svc.ServiceContext, avatarURL string) (*types.ChatAdminProfileResp, error) {
	l := chat_auth.NewUpdateProfileLogic(r.Context(), svcCtx)
	return l.UpdateProfile(&types.UpdateChatAdminProfileReq{
		AvatarUrl: avatarURL,
	})
}
