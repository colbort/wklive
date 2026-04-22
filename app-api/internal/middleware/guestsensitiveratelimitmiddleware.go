package middleware

import (
	"net/http"
	"strings"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type GuestSensitiveRateLimitMiddleware struct {
	rds *redis.Redis
}

func NewGuestSensitiveRateLimitMiddleware(rds *redis.Redis) *GuestSensitiveRateLimitMiddleware {
	return &GuestSensitiveRateLimitMiddleware{rds: rds}
}

func (m *GuestSensitiveRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetClientIP(r)

		// 每 IP 每秒 5 次，突发 10
		tokenLimiter := limit.NewTokenLimiter(5, 10, m.rds, utils.BuildIPKey("rl:guest_sensitive", ip))
		if !tokenLimiter.AllowCtx(r.Context()) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
				"code": 429,
				"msg":  "操作过于频繁",
			})
			return
		}

		// 每 IP 每 10 分钟最多 30 次
		periodLimiter := limit.NewPeriodLimit(600, 30, m.rds, "pl:guest_sensitive:ip")
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
				"msg":  "操作过于频繁",
			})
			return
		}

		// 可选：按手机号/邮箱/用户名进一步限制
		account := strings.TrimSpace(r.URL.Query().Get("account"))
		if account != "" {
			accLimiter := limit.NewPeriodLimit(600, 10, m.rds, "pl:guest_sensitive:account")
			code, err = accLimiter.Take(account)
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
					"msg":  "该账号操作过于频繁",
				})
				return
			}
		}

		next(w, r)
	}
}
