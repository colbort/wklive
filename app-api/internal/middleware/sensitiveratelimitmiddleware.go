package middleware

import (
	"net/http"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type SensitiveRateLimitMiddleware struct {
	rds *redis.Redis
}

func NewSensitiveRateLimitMiddleware(rds *redis.Redis) *SensitiveRateLimitMiddleware {
	return &SensitiveRateLimitMiddleware{rds: rds}
}

func (m *SensitiveRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
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

		// 1) 每 uid 每秒 3 次，突发 5
		tokenLimiter := limit.NewTokenLimiter(3, 5, m.rds, utils.BuildUidKey("rl:sensitive:user", uid))
		if !tokenLimiter.AllowCtx(r.Context()) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"code": 429,
				"msg":  "操作过于频繁",
			})
			return
		}

		// 2) 每 uid 每分钟最多 30 次
		periodLimiter := limit.NewPeriodLimit(60, 30, m.rds, "pl:sensitive:user")
		code, err := periodLimiter.Take(utils.BuildUidKey("user", uid))
		if err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, map[string]any{
				"code": 500,
				"msg":  "限流服务异常",
			})
			return
		}
		if code == limit.OverQuota {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"code": 429,
				"msg":  "操作过于频繁",
			})
			return
		}

		// 3) IP 兜底，每 IP 每秒 10 次，突发 20
		ipLimiter := limit.NewTokenLimiter(10, 20, m.rds, utils.BuildIPKey("rl:sensitive:ip", ip))
		if !ipLimiter.AllowCtx(r.Context()) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"code": 429,
				"msg":  "操作过于频繁",
			})
			return
		}

		next(w, r)
	}
}
