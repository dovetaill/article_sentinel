package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDGeneratesAndPropagatesContextValue(t *testing.T) {
	var captured string
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("request id missing from context")
		}
		captured = requestID
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if captured == "" {
		t.Fatal("captured request id = empty")
	}
	if got := rec.Header().Get(requestIDHeader); got == "" {
		t.Fatal("response header X-Request-ID is empty")
	} else if got != captured {
		t.Fatalf("response request id = %q, want %q", got, captured)
	}
}

func TestRequestIDPreservesInboundValue(t *testing.T) {
	const inboundRequestID = "req-123"

	var captured string
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("request id missing from context")
		}
		captured = requestID
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.Header.Set(requestIDHeader, inboundRequestID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if captured != inboundRequestID {
		t.Fatalf("context request id = %q, want %q", captured, inboundRequestID)
	}
	if got := rec.Header().Get(requestIDHeader); got != inboundRequestID {
		t.Fatalf("response request id = %q, want %q", got, inboundRequestID)
	}
}
