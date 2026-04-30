package payment

import (
	"net/http"

	"wklive/app-api/internal/logic/payment"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetMyCryptoRechargeAddressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMyCryptoRechargeAddressReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := payment.NewGetMyCryptoRechargeAddressLogic(r.Context(), svcCtx)
		resp, err := l.GetMyCryptoRechargeAddress(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
