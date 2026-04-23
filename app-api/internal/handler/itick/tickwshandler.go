// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"net/http"

	"wklive/app-api/internal/logic/itick"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TickWsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WsItickReq
		if err := httpx.Parse(r, &req); err != nil {
			logx.Errorf("获取参数失败 %s", err.Error())
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("tick ws upgrade failed: host=%s origin=%s err=%v",
				r.Host,
				r.Header.Get("Origin"),
				err,
			)
			httpx.Error(w, err)
			return
		}

		l := itick.NewTickWsLogic(r.Context(), svcCtx)
		if err := l.TickWs(conn, r); err != nil {
			logx.Errorf("tick ws logic closed: host=%s origin=%s err=%v",
				r.Host,
				r.Header.Get("Origin"),
				err,
			)
			return
		}
	}
}
