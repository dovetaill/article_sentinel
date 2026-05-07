package middleware

import (
	"net/http"
	"strings"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

func ProtectDocumentation(enabled bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enabled && isDocumentationPath(r.URL.Path) {
				if _, ok := identity.ActorFromContext(r.Context()); !ok {
					writeAuthError(w, http.StatusUnauthorized, "unauthorized")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isDocumentationPath(path string) bool {
	path = strings.TrimSpace(path)
	switch path {
	case "/docs", "/openapi.json", "/openapi.yaml", "/openapi-3.0.json", "/openapi-3.0.yaml":
		return true
	}
	return strings.HasPrefix(path, "/docs/") || strings.HasPrefix(path, "/schemas/")
}
