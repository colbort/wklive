// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"net/http"

	"wklive/admin-api/internal/logic/auth_private"
	"wklive/admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.Error(w, err)
			return
		}
		defer file.Close()

		l := auth_private.NewUploadAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadAvatar(file, header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
