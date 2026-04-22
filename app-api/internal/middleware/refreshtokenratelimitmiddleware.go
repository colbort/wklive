package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type RefreshTokenRateLimitMiddleware struct {
	rds *redis.Redis
}

func NewRefreshTokenRateLimitMiddleware(rds *redis.Redis) *RefreshTokenRateLimitMiddleware {
	return &RefreshTokenRateLimitMiddleware{rds: rds}
}

type refreshTokenBody struct {
	RefreshToken string `json:"refreshToken"`
}

func (m *RefreshTokenRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := utils.GetClientIP(r)

		refreshToken := extractRefreshToken(r)

		// 1) 优先按 refresh token 限流
		if refreshToken != "" {
			// 每个 refresh token 每秒 2 次，突发 3 次
			tokenLimiter := limit.NewTokenLimiter(
				2,
				3,
				m.rds,
				utils.BuildStringKey("rl:refresh_token", refreshToken),
			)
			if !tokenLimiter.AllowCtx(r.Context()) {
				httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]any{
					"code": 429,
					"msg":  "刷新过于频繁",
				})
				return
			}

			// 每个 refresh token 每分钟最多 20 次
			periodLimiter := limit.NewPeriodLimit(60, 20, m.rds, "pl:refresh_token")
			code, err := periodLimiter.Take(refreshToken)
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
					"msg":  "刷新过于频繁",
				})
				return
			}
		}

		// 2) 再按 IP 兜底
		ipLimiter := limit.NewPeriodLimit(60, 120, m.rds, "pl:refresh_token:ip")
		code, err := ipLimiter.Take(ip)
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
				"msg":  "刷新过于频繁",
			})
			return
		}

		next(w, r)
	}
}

func extractRefreshToken(r *http.Request) string {
	// 1. 先从 Header 取
	if v := strings.TrimSpace(r.Header.Get("X-Refresh-Token")); v != "" {
		return v
	}

	// 2. 再从 Authorization 取（如果你项目这么传）
	if v := strings.TrimSpace(r.Header.Get("Authorization")); v != "" {
		// 例如: Bearer xxxxx
		if strings.HasPrefix(strings.ToLower(v), "bearer ") {
			return strings.TrimSpace(v[7:])
		}
		return v
	}

	// 3. 再从 JSON body 取
	if r.Body == nil {
		return ""
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	// 读完后要塞回去，不然下游 handler 读不到
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	if len(bodyBytes) == 0 {
		return ""
	}

	var req refreshTokenBody
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return ""
	}

	return strings.TrimSpace(req.RefreshToken)
}
