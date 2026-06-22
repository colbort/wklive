// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"context"
	"net/http"
	"strconv"

	"wklive/common/utils"
)

type AdminRateLimitMiddleware struct {
}

func NewAdminRateLimitMiddleware() *AdminRateLimitMiddleware {
	return &AdminRateLimitMiddleware{}
}

func (m *AdminRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if merchantID := merchantIDFromHeader(r); merchantID > 0 {
			r = r.WithContext(context.WithValue(r.Context(), utils.CtxKeyMerchantId, merchantID))
		}

		// Passthrough to next handler if need
		next(w, r)
	}
}

func merchantIDFromHeader(r *http.Request) int64 {
	if r == nil {
		return 0
	}
	for _, key := range []string{utils.CtxKeyMerchantId, "merchantId"} {
		if merchantID, err := strconv.ParseInt(r.Header.Get(key), 10, 64); err == nil && merchantID > 0 {
			return merchantID
		}
	}
	return 0
}
