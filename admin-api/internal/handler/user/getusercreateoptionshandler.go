package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"wklive/admin-api/internal/logic/user"
	"wklive/admin-api/internal/svc"
)

func GetUserCreateOptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetUserCreateOptionsLogic(r.Context(), svcCtx)
		resp, err := l.GetUserCreateOptions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
