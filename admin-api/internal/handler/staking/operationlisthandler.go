// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/admin-api/internal/logic/staking"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
)

func OperationListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OperationListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := staking.NewOperationListLogic(r.Context(), svcCtx)
		resp, err := l.OperationList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
