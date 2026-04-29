package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
	"github.com/dovetaill/article-sentinel/pkg/config"
)

func TestSessionContextStoresAdminSessionFromCookie(t *testing.T) {
	mgr := newTestAdminSessionManager()
	token := signedSessionToken(t, mgr, identity.AdminSession{
		UserID:   90525,
		OrgID:    29,
		Nickname: "用户A",
		Priv:     "super",
	})

	handler := SessionContext(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := identity.AdminSessionFromContext(r.Context())
		if !ok || session.OrgID != 29 {
			t.Fatalf("session = %+v, ok=%v", session, ok)
		}
		actor, ok := identity.ActorFromContext(r.Context())
		if !ok || actor.Username != "用户A" {
			t.Fatalf("actor = %+v, ok=%v", actor, ok)
		}
		principal, ok := identity.PrincipalFromContext(r.Context())
		if !ok || principal.UserID != 90525 {
			t.Fatalf("principal = %+v, ok=%v", principal, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/article-inspect/tasks", nil)
	req.AddCookie(&http.Cookie{Name: mgr.CookieName(), Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSessionContextClearsInvalidCookie(t *testing.T) {
	mgr := newTestAdminSessionManager()
	handler := SessionContext(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	req.AddCookie(&http.Cookie{Name: mgr.CookieName(), Value: "bad-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") || !strings.Contains(got, "Max-Age=0") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func TestSessionContextAllowsAnonymousRequestWithoutCookie(t *testing.T) {
	mgr := newTestAdminSessionManager()
	handler := SessionContext(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := identity.AdminSessionFromContext(r.Context()); ok {
			t.Fatal("unexpected admin session in anonymous request")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func newTestAdminSessionManager() *identity.AdminSessionManager {
	return identity.NewAdminSessionManager(config.SessionConfig{
		LegacySecret: "legacy-secret",
		Secret:       "session-secret",
		Issuer:       "article-sentinel-admin",
		TTLHours:     24,
	})
}

func signedSessionToken(t *testing.T, mgr *identity.AdminSessionManager, session identity.AdminSession) string {
	t.Helper()

	token, _, err := mgr.SignSessionJWT(session)
	if err != nil {
		t.Fatalf("SignSessionJWT() error = %v", err)
	}
	return token
}
