// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package callback

import (
	"net/http"

	"wklive/payment-api/internal/logic/callback"
	"wklive/payment-api/internal/svc"
	"wklive/payment-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PayoutNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.NotifyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := callback.NewPayoutNotifyLogic(r.Context(), svcCtx)
		resp, err := l.PayoutNotify(handleNotify(w, r, svcCtx, &req))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
