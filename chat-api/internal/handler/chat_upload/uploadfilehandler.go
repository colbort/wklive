package chat_upload

import (
	"net/http"

	"chat-api/internal/logic/chat_upload"
	"chat-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, chat_upload.MaxChatUploadSize)
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := chat_upload.NewUploadFileLogic(r.Context(), svcCtx)
		resp, err := l.UploadFile(file, header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func DownloadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := chat_upload.NewUploadFileLogic(r.Context(), svcCtx)
		path, err := l.UploadedFilePath(r.URL.Query().Get("url"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		http.ServeFile(w, r, path)
	}
}
