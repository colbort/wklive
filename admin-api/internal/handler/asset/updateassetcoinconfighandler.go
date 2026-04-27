// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package asset

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/admin-api/internal/logic/asset"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
)

func UpdateAssetCoinConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateAssetCoinConfigReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := asset.NewUpdateAssetCoinConfigLogic(r.Context(), svcCtx)
		resp, err := l.UpdateAssetCoinConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
