// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/admin-api/internal/logic/itick"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
)

func CreateTenantCategoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateTenantCategoryReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := itick.NewCreateTenantCategoryLogic(r.Context(), svcCtx)
		resp, err := l.CreateTenantCategory(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
