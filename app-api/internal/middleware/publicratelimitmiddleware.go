package middleware

import (
	"net/http"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type PublicRateLimitMiddleware struct {
	rds *redis.Redis
}

func NewPublicRateLimitMiddleware(rds *redis.Redis) *PublicRateLimitMiddleware {
	return &PublicRateLimitMiddleware{rds: rds}
}

func (m *PublicRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetClientIP(r)

		// 每 IP 每秒 20 次，突发 40
		tokenLimiter := limit.NewTokenLimiter(20, 40, m.rds, utils.BuildIPKey("rl:public", ip))
		if !tokenLimiter.AllowCtx(r.Context()) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"code": 429,
				"msg":  "请求过于频繁",
			})
			return
		}

		// 每 IP 每分钟最多 300 次
		periodLimiter := limit.NewPeriodLimit(60, 300, m.rds, "pl:public")
		code, err := periodLimiter.Take(ip)
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
				"msg":  "请求过于频繁",
			})
			return
		}

		next(w, r)
	}
}