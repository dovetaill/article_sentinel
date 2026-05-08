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

func TestAuthExchangeMintsOneTimeCodeAndLoginConsumesIt(t *testing.T) {
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

	exchangeReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/exchange",
		strings.NewReader(url.Values{"jwt": []string{legacy}}.Encode()),
	)
	exchangeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	exchangeRec := httptest.NewRecorder()
	handler.ServeHTTP(exchangeRec, exchangeReq)

	if exchangeRec.Code != http.StatusOK {
		t.Fatalf("exchange status = %d, want %d", exchangeRec.Code, http.StatusOK)
	}

	var exchangeEnvelope response.Envelope
	if err := json.Unmarshal(exchangeRec.Body.Bytes(), &exchangeEnvelope); err != nil {
		t.Fatalf("decode exchange response: %v", err)
	}

	data, ok := exchangeEnvelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("exchange data type = %T, want map[string]any", exchangeEnvelope.Data)
	}

	code, _ := data["code"].(string)
	if code == "" {
		t.Fatalf("exchange code = %q, want non-empty", code)
	}

	loginReq := httptest.NewRequest(http.MethodGet, "/auth/login?code="+url.QueryEscape(code), nil)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusFound {
		t.Fatalf("login status = %d, want %d", loginRec.Code, http.StatusFound)
	}
	if got := loginRec.Header().Get("Location"); got != "http://127.0.0.1:5173/" {
		t.Fatalf("Location = %q, want %q", got, "http://127.0.0.1:5173/")
	}
	if got := loginRec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func TestAuthLoginBridgesLegacyQueryJWTByDefault(t *testing.T) {
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
	if got := rec.Header().Get("Location"); got != "http://127.0.0.1:5173/" {
		t.Fatalf("Location = %q, want %q", got, "http://127.0.0.1:5173/")
	}
	if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") || strings.Contains(got, "Max-Age=0") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func TestAuthLoginBridgesLegacyQueryJWTWithConfiguredURLs(t *testing.T) {
	handler, mgr := newAuthHandlerForTestWithConfig(t, config.SessionConfig{
		LoginURL:    "https://appadmin.cq.qiludev.com/cq-admin/index.html#/home",
		RedirectURL: "http://127.0.0.1:5173/",
	})
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
	if got := rec.Header().Get("Location"); got != "http://127.0.0.1:5173/" {
		t.Fatalf("Location = %q, want %q", got, "http://127.0.0.1:5173/")
	}
	if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func TestAuthLoginCodeIsSingleUse(t *testing.T) {
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

	code := exchangeLegacyJWTForCode(t, handler, legacy)
	firstReq := httptest.NewRequest(http.MethodGet, "/auth/login?code="+url.QueryEscape(code), nil)
	firstRec := httptest.NewRecorder()
	handler.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusFound {
		t.Fatalf("first status = %d, want %d", firstRec.Code, http.StatusFound)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/auth/login?code="+url.QueryEscape(code), nil)
	secondRec := httptest.NewRecorder()
	handler.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusFound {
		t.Fatalf("second status = %d, want %d", secondRec.Code, http.StatusFound)
	}
	if got := secondRec.Header().Get("Location"); got != "https://appadmin.cq.qiludev.com/cq-admin/index.html#/home" {
		t.Fatalf("Location = %q, want %q", got, "https://appadmin.cq.qiludev.com/cq-admin/index.html#/home")
	}
	if got := secondRec.Header().Get("Set-Cookie"); !strings.Contains(got, mgr.CookieName()+"=") || !strings.Contains(got, "Max-Age=0") {
		t.Fatalf("Set-Cookie = %q", got)
	}
}

func TestAuthLoginSetsNoStoreHeaders(t *testing.T) {
	handler, _ := newAuthHandlerForTest(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "no-store, max-age=0" {
		t.Fatalf("Cache-Control = %q, want %q", got, "no-store, max-age=0")
	}
	if got := rec.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("Pragma = %q, want %q", got, "no-cache")
	}
	if got := rec.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("Referrer-Policy = %q, want %q", got, "no-referrer")
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
	return newAuthHandlerForTestWithConfig(t, config.SessionConfig{
		LoginURL:    "https://appadmin.cq.qiludev.com/cq-admin/index.html#/home",
		RedirectURL: "http://127.0.0.1:5173/",
	})
}

func newAuthHandlerForTestWithConfig(t *testing.T, sessionCfg config.SessionConfig) (http.Handler, *identity.AdminSessionManager) {
	t.Helper()

	mux := http.NewServeMux()
	manager := identity.NewAdminSessionManager(config.SessionConfig{
		LegacySecret: "legacy-secret",
		Secret:       "session-secret",
		Issuer:       "article-sentinel-admin",
		TTLHours:     24,
		LoginURL:     "https://appadmin.cq.qiludev.com/cq-admin/index.html#/home",
		RedirectURL:  "http://127.0.0.1:5173/",
	})
	handlers.RegisterAuthRoutes(mux, manager, sessionCfg)
	return middleware.Chain(mux, middleware.SessionContext(manager)), manager
}

func exchangeLegacyJWTForCode(t *testing.T, handler http.Handler, legacy string) string {
	t.Helper()

	exchangeReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/exchange",
		strings.NewReader(url.Values{"jwt": []string{legacy}}.Encode()),
	)
	exchangeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	exchangeRec := httptest.NewRecorder()
	handler.ServeHTTP(exchangeRec, exchangeReq)

	if exchangeRec.Code != http.StatusOK {
		t.Fatalf("exchange status = %d, want %d", exchangeRec.Code, http.StatusOK)
	}

	var exchangeEnvelope response.Envelope
	if err := json.Unmarshal(exchangeRec.Body.Bytes(), &exchangeEnvelope); err != nil {
		t.Fatalf("decode exchange response: %v", err)
	}

	data, ok := exchangeEnvelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("exchange data type = %T, want map[string]any", exchangeEnvelope.Data)
	}

	code, _ := data["code"].(string)
	if code == "" {
		t.Fatal("exchange code = empty")
	}

	return code
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
