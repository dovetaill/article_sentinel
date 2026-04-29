package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/api/handlers"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
	"github.com/dovetaill/article-sentinel/internal/middleware"
	"github.com/dovetaill/article-sentinel/pkg/config"
	jwt "github.com/golang-jwt/jwt/v5"
)

func TestAuthLoginBridgesLegacyJWTIntoSessionCookie(t *testing.T) {
	handler, mgr := newAuthHandlerForTest(t)
	legacy := signedLegacyJWTForAuthTest(t, mgr, map[string]any{
		"id":       "90525",
		"orgid":    "29",
		"orgname":  "一县一端测试机构",
		"nickname": "用户A",
		"avatar":   "https://example.com/a.png",
		"platform": "chuangqi",
		"priv":     "super",
		"roleid":   "1",
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/login?jwt="+url.QueryEscape(legacy), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "/" {
		t.Fatalf("Location = %q, want %q", got, "/")
	}
	if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func TestAuthLoginClearsCookieAndRedirectsToFixedLoginOnInvalidJWT(t *testing.T) {
	handler, mgr := newAuthHandlerForTest(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/login?jwt=bad-token", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != handlers.FixedAdminLoginURL {
		t.Fatalf("Location = %q, want %q", got, handlers.FixedAdminLoginURL)
	}
	if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") || !strings.Contains(got, "Max-Age=0") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func TestAuthSessionReturnsCurrentSessionEnvelope(t *testing.T) {
	handler, mgr := newAuthHandlerForTest(t)
	token := signedSessionJWTForAuthTest(t, mgr, identity.AdminSession{
		UserID:       90525,
		OrgID:        29,
		OrgName:      "一县一端测试机构",
		Nickname:     "用户A",
		Avatar:       "https://example.com/a.png",
		Platform:     "chuangqi",
		Priv:         "super",
		RoleID:       "1",
		DepartmentID: 3,
		IsOpenEdu:    true,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	req.AddCookie(&http.Cookie{Name: mgr.CookieName(), Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", got.Data)
	}
	if data["orgid"] != float64(29) || data["nickname"] != "用户A" {
		t.Fatalf("session data = %#v", data)
	}
}

func TestAuthSessionReturnsUnauthorizedAndClearsCookieWhenCookieInvalid(t *testing.T) {
	handler, mgr := newAuthHandlerForTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	req.AddCookie(&http.Cookie{Name: mgr.CookieName(), Value: "bad-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") || !strings.Contains(got, "Max-Age=0") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func TestAuthLogoutAlwaysClearsCookie(t *testing.T) {
	handler, mgr := newAuthHandlerForTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") || !strings.Contains(got, "Max-Age=0") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func newAuthHandlerForTest(t *testing.T) (http.Handler, *identity.AdminSessionManager) {
	t.Helper()

	mux := http.NewServeMux()
	manager := identity.NewAdminSessionManager(config.SessionConfig{
		LegacySecret: "legacy-secret",
		Secret:       "session-secret",
		CookieName:   "as_admin_session",
		Issuer:       "article-sentinel-admin",
		TTLHours:     24,
	})
	handlers.RegisterAuthRoutes(mux, manager)
	return middleware.Chain(mux, middleware.SessionContext(manager)), manager
}

func signedLegacyJWTForAuthTest(t *testing.T, mgr *identity.AdminSessionManager, claims map[string]any) string {
	t.Helper()

	_ = mgr
	token := signedMapClaimsForAuthTest(t, claims, []byte("legacy-secret"))
	return token
}

func signedSessionJWTForAuthTest(t *testing.T, mgr *identity.AdminSessionManager, session identity.AdminSession) string {
	t.Helper()

	token, _, err := mgr.SignSessionJWT(session)
	if err != nil {
		t.Fatalf("SignSessionJWT() error = %v", err)
	}
	return token
}

func signedMapClaimsForAuthTest(t *testing.T, claims map[string]any, secret []byte) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}
