package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/go-auth-demo/internal/api/handlers"
	"github.com/dovetaill/go-auth-demo/internal/api/response"
	"github.com/dovetaill/go-auth-demo/internal/identity"
)

func TestDemoMeRequiresAuthenticatedActor(t *testing.T) {
	handler := newDemoHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/demo/me", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want %d", got.Code, http.StatusUnauthorized)
	}
	if got.Message != "unauthorized" {
		t.Fatalf("message = %q, want %q", got.Message, "unauthorized")
	}
}

func TestDemoMeReturnsCurrentActorFromContext(t *testing.T) {
	handler := newDemoHandler()
	actor := identity.NewActor(7, "demo", "admin", "active")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/demo/me", nil)
	req = req.WithContext(identity.ContextWithActor(req.Context(), actor))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != 0 {
		t.Fatalf("code = %d, want %d", got.Code, 0)
	}
	if got.Message != "me" {
		t.Fatalf("message = %q, want %q", got.Message, "me")
	}

	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", got.Data)
	}
	if got := data["id"]; got != float64(7) {
		t.Fatalf("data.id = %v, want %d", got, 7)
	}
	if got := data["username"]; got != "demo" {
		t.Fatalf("data.username = %v, want %q", got, "demo")
	}
	if got := data["role"]; got != "admin" {
		t.Fatalf("data.role = %v, want %q", got, "admin")
	}
	if got := data["status"]; got != "active" {
		t.Fatalf("data.status = %v, want %q", got, "active")
	}
}

func newDemoHandler() http.Handler {
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	handlers.RegisterDemoRoutes(api)
	return mux
}
