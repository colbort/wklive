// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_public

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/app-api/internal/logic/user_public"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
)

func ExchangeGuestTransferHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ExchangeGuestTransferReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		currentOrigin := r.Header.Get("Origin")

		l := user_public.NewExchangeGuestTransferLogic(r.Context(), svcCtx)
		resp, err := l.ExchangeGuestTransfer(&req, currentOrigin)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
