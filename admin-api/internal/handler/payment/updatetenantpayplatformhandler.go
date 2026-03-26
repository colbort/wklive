// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/admin-api/internal/logic/payment"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
)

func UpdateTenantPayPlatformHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateTenantPayPlatformReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := payment.NewUpdateTenantPayPlatformLogic(r.Context(), svcCtx)
		resp, err := l.UpdateTenantPayPlatform(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
