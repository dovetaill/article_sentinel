package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

type stubAuthenticator struct {
	actor *identity.Actor
	err   error
}

func (s stubAuthenticator) Authenticate(ctx context.Context, token string) (*identity.Actor, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.actor, nil
}

func TestAuthenticateStoresGenericActor(t *testing.T) {
	handler := Authenticate(stubAuthenticator{
		actor: &identity.Actor{
			ID:       1,
			Username: "editor",
			Role:     "editor",
			Status:   "active",
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := identity.ActorFromContext(r.Context())
		if !ok {
			t.Fatal("actor missing from context")
		}
		if actor.Role != "editor" {
			t.Fatalf("actor role = %q, want %q", actor.Role, "editor")
		}
		principal, ok := identity.PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("principal missing from context")
		}
		if principal.Kind != identity.PrincipalKind("editor") {
			t.Fatalf("principal kind = %q, want %q", principal.Kind, identity.PrincipalKind("editor"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireRoleAllowsMatchingActor(t *testing.T) {
	handler := Authenticate(stubAuthenticator{
		actor: &identity.Actor{
			ID:       9,
			Username: "writer",
			Role:     "editor",
			Status:   "active",
		},
	})(RequireRole("editor")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireAdminRejectsNonAdminActor(t *testing.T) {
	handler := Authenticate(stubAuthenticator{
		actor: &identity.Actor{
			ID:       11,
			Username: "writer",
			Role:     "editor",
			Status:   "active",
		},
	})(RequireAdmin()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRequireAuthenticatedRejectsAnonymous(t *testing.T) {
	handler := RequireAuthenticated()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
