package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

const requestIDHeader = "X-Request-ID"

var requestIDCounter uint64

type requestIDContextKey struct{}

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

			ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
			w.Header().Set(requestIDHeader, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext 返回上下文中的 request id。
func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	if !ok || requestID == "" {
		return "", false
	}
	return requestID, true
}
