package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

func TestRequestMetadataUsesTrustedProxyForwardedAddress(t *testing.T) {
	var captured identity.RequestMetadata

	handler := RequestMetadata([]string{"127.0.0.1/32"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := identity.RequestMetadataFromContext(r.Context())
		if !ok {
			t.Fatal("request metadata missing from context")
		}
		captured = metadata
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("X-Forwarded-For", "203.0.113.8, 127.0.0.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if captured.SourceIP != "203.0.113.8" {
		t.Fatalf("SourceIP = %q, want %q", captured.SourceIP, "203.0.113.8")
	}
}
