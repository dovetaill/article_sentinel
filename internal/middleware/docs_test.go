package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dovetaill/go-auth-demo/internal/api/response"
	"github.com/dovetaill/go-auth-demo/internal/identity"
)

func TestProtectDocumentationRejectsAnonymousDocumentationPaths(t *testing.T) {
	protectedPaths := []string{
		"/docs",
		"/docs/",
		"/docs/index.html",
		"/openapi.json",
		"/openapi.yaml",
		"/openapi-3.0.json",
		"/openapi-3.0.yaml",
		"/schemas/ErrorModel.json",
	}

	for _, path := range protectedPaths {
		t.Run(path, func(t *testing.T) {
			handler := ProtectDocumentation(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatalf("next handler should not run for anonymous request to %s", path)
			}))

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
			if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
				t.Fatalf("Content-Type = %q, want application/json", contentType)
			}

			var got response.Envelope
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Code != http.StatusUnauthorized || got.Message != "unauthorized" {
				t.Fatalf("body = %+v, want unauthorized auth envelope", got)
			}
		})
	}
}

func TestProtectDocumentationAllowsActorContext(t *testing.T) {
	protectedPaths := []string{
		"/docs",
		"/openapi.json",
		"/schemas/ErrorModel.json",
	}

	for _, path := range protectedPaths {
		t.Run(path, func(t *testing.T) {
			handler := ProtectDocumentation(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, ok := identity.ActorFromContext(r.Context()); !ok {
					t.Fatal("actor missing from protected documentation request")
				}
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, path, nil)
			ctx := identity.ContextWithActor(req.Context(), identity.NewActor(1, "admin", "admin", "active"))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req.WithContext(ctx))

			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
			}
		})
	}
}

func TestProtectDocumentationAllowsAnonymousNonDocumentationPaths(t *testing.T) {
	handler := ProtectDocumentation(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestProtectDocumentationDisabledAllowsDocumentationPaths(t *testing.T) {
	handler := ProtectDocumentation(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
