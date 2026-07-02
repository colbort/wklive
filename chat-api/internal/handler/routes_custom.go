package handler

import (
	"net/http"

	"chat-api/internal/handler/chat_upload"
	"chat-api/internal/handler/chat_ws"
	"chat-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterCustomHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.UserRateLimit},
			[]rest.Route{
				{
					Method:  http.MethodGet,
					Path:    "/ws/messages",
					Handler: chat_ws.MessagesHandler(serverCtx),
				},
			}...,
		),
		rest.WithPrefix("/chat"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.UserRateLimit},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/upload/file",
					Handler: chat_upload.UploadFileHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/upload/file",
					Handler: chat_upload.DownloadFileHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Jwt.AccessSecret),
		rest.WithPrefix("/chat"),
	)
}
