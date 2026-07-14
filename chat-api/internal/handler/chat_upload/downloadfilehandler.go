package chat_upload

import (
	"net/http"

	"chat-api/internal/logic/chat_upload"
	"chat-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DownloadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := chat_upload.NewDownloadFileLogic(r.Context(), svcCtx)
		path, err := l.DownloadedFilePath(r.URL.Query().Get("url"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			http.ServeFile(w, r, path)
		}
	}
}
