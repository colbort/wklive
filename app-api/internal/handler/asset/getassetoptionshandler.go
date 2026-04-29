package asset

import (
	"net/http"

	"wklive/app-api/internal/logic/asset"
	"wklive/app-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAssetOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := asset.NewGetAssetOptionsLogic(r.Context(), svcCtx)
		resp, err := l.GetAssetOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
