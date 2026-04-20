package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusCapturingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusCapturingResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(p)
}

func (w *statusCapturingResponseWriter) status() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

// AccessLog 记录请求访问日志。
func AccessLog(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			writer := &statusCapturingResponseWriter{ResponseWriter: w}
			next.ServeHTTP(writer, r)
			if logger != nil {
				attrs := []any{
					"method", r.Method,
					"path", r.URL.Path,
					"status_code", writer.status(),
					"duration_ms", time.Since(start).Milliseconds(),
				}
				if requestID, ok := RequestIDFromContext(r.Context()); ok {
					attrs = append(attrs, "request_id", requestID)
				}
				logger.Info("http_access", attrs...)
			}
		})
	}
}
