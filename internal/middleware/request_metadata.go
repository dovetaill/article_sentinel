package middleware

import (
	"net/http"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

// RequestMetadata 注入可信来源解析后的 source ip，供审计和操作日志复用。
func RequestMetadata(trustedProxyCIDRs []string) Middleware {
	trustedProxyNets := parseTrustedProxyCIDRs(trustedProxyCIDRs)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := identity.ContextWithRequestMetadata(r.Context(), identity.RequestMetadata{
				SourceIP: clientIPFromRequest(r, trustedProxyNets),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
