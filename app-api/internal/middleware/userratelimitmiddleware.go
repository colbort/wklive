package middleware

import (
	"net/http"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type UserRateLimitMiddleware struct {
	rds *redis.Redis
}

func NewUserRateLimitMiddleware(rds *redis.Redis) *UserRateLimitMiddleware {
	return &UserRateLimitMiddleware{rds: rds}
}

func (m *UserRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := utils.GetUidFromCtx(r.Context())
		if err != nil || uid <= 0 {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, map[string]any{
				"code": 401,
				"msg":  "未登录",
			})
			return
		}

		ip := utils.GetClientIP(r)

		// 每 uid 每秒 10 次，突发 20
		userLimiter := limit.NewTokenLimiter(10, 20, m.rds, utils.BuildUidKey("rl:user", uid))
		if !userLimiter.AllowCtx(r.Context()) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"code": 429,
				"msg":  "请求过于频繁",
			})
			return
		}

		// IP 兜底
		ipLimiter := limit.NewTokenLimiter(30, 60, m.rds, utils.BuildIPKey("rl:user_ip", ip))
		if !ipLimiter.AllowCtx(r.Context()) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"code": 429,
				"msg":  "请求过于频繁",
			})
			return
		}

		next(w, r)
	}
}
