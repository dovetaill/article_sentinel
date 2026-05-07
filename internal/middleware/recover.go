package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/dovetaill/go-auth-demo/internal/api/response"
)

// Recover 捕获 panic，避免服务崩溃。
func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(response.Fail(http.StatusInternalServerError, "internal server error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
