// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/user_private"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
)

func CreateGuestTransferHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateGuestTransferReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		sourceOrigin := r.Header.Get("Origin")

		l := user_private.NewCreateGuestTransferLogic(r.Context(), svcCtx)
		resp, err := l.CreateGuestTransfer(&req, sourceOrigin)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
