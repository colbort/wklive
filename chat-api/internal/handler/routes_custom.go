package handler

import (
	"net/http"

	chat_ws "chat-api/internal/handler/chat_ws"
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
}
