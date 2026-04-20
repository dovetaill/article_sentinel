package middleware

import (
	"net/http"
	"time"
)

// Timeout 为请求附加超时控制。
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		if timeout <= 0 {
			return next
		}
		return http.TimeoutHandler(next, timeout, "request timeout")
	}
}
