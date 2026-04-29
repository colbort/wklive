package payment

import (
	"net/http"

	"wklive/app-api/internal/logic/payment"
	"wklive/app-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPaymentOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := payment.NewGetPaymentOptionsLogic(r.Context(), svcCtx)
		resp, err := l.GetPaymentOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
