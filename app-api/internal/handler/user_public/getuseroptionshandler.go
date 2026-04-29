package user_public

import (
	"net/http"

	"wklive/app-api/internal/logic/user_public"
	"wklive/app-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetUserOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user_public.NewGetUserOptionsLogic(r.Context(), svcCtx)
		resp, err := l.GetUserOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
