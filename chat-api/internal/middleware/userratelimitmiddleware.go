// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import "net/http"

type UserRateLimitMiddleware struct {
}

func NewUserRateLimitMiddleware() *UserRateLimitMiddleware {
	return &UserRateLimitMiddleware{}
}

func (m *UserRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need
		next(w, r)
	}
}
