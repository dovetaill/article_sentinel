package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

const requestIDHeader = "X-Request-ID"

var requestIDCounter uint64

// Middleware 表示 HTTP 中间件。
type Middleware func(http.Handler) http.Handler

// Chain 将中间件按传入顺序应用到 handler。
func Chain(handler http.Handler, mws ...Middleware) http.Handler {
	wrapped := handler
	for i := len(mws) - 1; i >= 0; i-- {
		if mws[i] == nil {
			continue
		}
		wrapped = mws[i](wrapped)
	}
	return wrapped
}

// RequestID 注入请求 ID 到上下文与响应头。
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + strconv.FormatUint(atomic.AddUint64(&requestIDCounter, 1), 10)
			}

			ctx := identity.ContextWithRequestMetadata(r.Context(), identity.RequestMetadata{RequestID: requestID})
			w.Header().Set(requestIDHeader, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext 返回上下文中的 request id。
func RequestIDFromContext(ctx context.Context) (string, bool) {
	metadata, ok := identity.RequestMetadataFromContext(ctx)
	if !ok || metadata.RequestID == "" {
		return "", false
	}
	return metadata.RequestID, true
}
