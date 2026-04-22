package middleware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type RequestLogMiddleware struct {
}

func NewRequestLogMiddleware() *RequestLogMiddleware {
	return &RequestLogMiddleware{}
}

func (m *RequestLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Get("")
		start := time.Now()

		logx.Infof(
			"[APP-API REQUEST] method=%s uri=%s host=%s remote=%s origin=%s ua=%s upgrade=%s connection=%s",
			r.Method,
			r.RequestURI,
			r.Host,
			r.RemoteAddr,
			r.Header.Get("Origin"),
			r.UserAgent(),
			r.Header.Get("Upgrade"),
			r.Header.Get("Connection"),
		)

		next(w, r)

		logx.Infof(
			"[APP-API REQUEST DONE] method=%s uri=%s cost=%s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}
}
