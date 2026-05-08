package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
	"github.com/dovetaill/article-sentinel/pkg/config"
)

const (
	DefaultAdminLoginURL    = "https://appadmin.cq.qiludev.com/cq-admin/index.html"
	DefaultAdminRedirectURL = "/"
	defaultExchangeCodeTTL  = 90 * time.Second
)

type AuthHandler struct {
	manager             *identity.AdminSessionManager
	loginURL            string
	redirectURL         string
	allowLegacyQueryJWT bool
	exchangeCodeStore   *authExchangeCodeStore
}

func RegisterAuthRoutes(mux *http.ServeMux, manager *identity.AdminSessionManager, sessionCfg config.SessionConfig) {
	if mux == nil {
		return
	}

	handler := &AuthHandler{
		manager:             manager,
		loginURL:            normalizeRedirectTarget(sessionCfg.LoginURL, DefaultAdminLoginURL),
		redirectURL:         normalizeRedirectTarget(sessionCfg.RedirectURL, DefaultAdminRedirectURL),
		allowLegacyQueryJWT: sessionCfg.AllowLegacyQueryJWT,
		exchangeCodeStore:   newAuthExchangeCodeStore(defaultExchangeCodeTTL),
	}
	mux.Handle("/auth/login", http.HandlerFunc(handler.Login))
	mux.Handle("/api/v1/auth/exchange", http.HandlerFunc(handler.Exchange))
	mux.Handle("/api/v1/auth/session", http.HandlerFunc(handler.Session))
	mux.Handle("/api/v1/auth/logout", http.HandlerFunc(handler.Logout))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	setAuthNoStoreHeaders(w)
	if h == nil || h.manager == nil {
		clearAdminSessionCookie(w, nil)
		http.Redirect(w, r, h.loginRedirectURL(), http.StatusFound)
		return
	}

	if code := strings.TrimSpace(r.URL.Query().Get("code")); code != "" {
		entry, ok := h.exchangeCodeStore.Consume(code)
		if !ok {
			clearAdminSessionCookie(w, h.manager)
			http.Redirect(w, r, h.loginRedirectURL(), http.StatusFound)
			return
		}
		setAdminSessionCookie(w, h.manager, entry.sessionJWT, entry.expiresAt)
		http.Redirect(w, r, h.successRedirectURL(), http.StatusFound)
		return
	}

	legacyJWT := strings.TrimSpace(r.URL.Query().Get("jwt"))
	if legacyJWT == "" {
		clearAdminSessionCookie(w, h.manager)
		http.Redirect(w, r, h.loginRedirectURL(), http.StatusFound)
		return
	}
	if !h.allowLegacyQueryJWT {
		clearAdminSessionCookie(w, h.manager)
		http.Redirect(w, r, h.loginRedirectURL(), http.StatusFound)
		return
	}

	sessionJWT, _, expiresAt, err := h.manager.ExchangeLegacyJWT(legacyJWT)
	if err != nil {
		clearAdminSessionCookie(w, h.manager)
		http.Redirect(w, r, h.loginRedirectURL(), http.StatusFound)
		return
	}

	setAdminSessionCookie(w, h.manager, sessionJWT, expiresAt)
	http.Redirect(w, r, h.successRedirectURL(), http.StatusFound)
}

func (h *AuthHandler) Exchange(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	setAuthNoStoreHeaders(w)
	if h == nil || h.manager == nil || h.exchangeCodeStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, response.Fail(http.StatusServiceUnavailable, "auth exchange unavailable"))
		return
	}

	legacyJWT, err := readLegacyJWTFromExchangeRequest(r)
	if err != nil {
		clearAdminSessionCookie(w, h.manager)
		writeJSON(w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, "legacy jwt is required"))
		return
	}

	sessionJWT, _, expiresAt, err := h.manager.ExchangeLegacyJWT(legacyJWT)
	if err != nil {
		clearAdminSessionCookie(w, h.manager)
		writeJSON(w, http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "unauthorized"))
		return
	}

	code, ttl, err := h.exchangeCodeStore.Issue(sessionJWT, expiresAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "auth exchange unavailable"))
		return
	}

	writeJSON(w, http.StatusOK, response.OK("exchange", map[string]any{
		"code":               code,
		"redirect_path":      "/auth/login?code=" + url.QueryEscape(code),
		"expires_in_seconds": int(ttl.Seconds()),
	}))
}

func (h *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	session, ok := identity.AdminSessionFromContext(r.Context())
	if !ok || session.OrgID == 0 {
		writeJSON(w, http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, "unauthorized"))
		return
	}

	writeJSON(w, http.StatusOK, response.OK("session", map[string]any{
		"id":           session.UserID,
		"orgid":        session.OrgID,
		"orgname":      session.OrgName,
		"platform":     session.Platform,
		"priv":         session.Priv,
		"roleid":       session.RoleID,
		"nickname":     session.Nickname,
		"avatar":       session.Avatar,
		"departmentid": session.DepartmentID,
		"is_open_edu":  session.IsOpenEdu,
	}))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	clearAdminSessionCookie(w, h.manager)
	writeJSON(w, http.StatusOK, response.OK("logout", map[string]any{
		"ok": true,
	}))
}

func writeJSON(w http.ResponseWriter, status int, body response.Envelope) {
	if w == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, response.Fail(http.StatusMethodNotAllowed, "method not allowed"))
}

func clearAdminSessionCookie(w http.ResponseWriter, manager *identity.AdminSessionManager) {
	name := "as_admin_session"
	secure := false
	if manager != nil {
		name = manager.CookieName()
		secure = manager.SecureCookie()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func setAdminSessionCookie(w http.ResponseWriter, manager *identity.AdminSessionManager, sessionJWT string, expiresAt time.Time) {
	if w == nil || manager == nil {
		return
	}

	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	http.SetCookie(w, &http.Cookie{
		Name:     manager.CookieName(),
		Value:    sessionJWT,
		Path:     "/",
		HttpOnly: true,
		Secure:   manager.SecureCookie(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
		Expires:  expiresAt,
	})
}

func setAuthNoStoreHeaders(w http.ResponseWriter) {
	if w == nil {
		return
	}
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

func (h *AuthHandler) loginRedirectURL() string {
	if h == nil {
		return DefaultAdminLoginURL
	}
	return normalizeRedirectTarget(h.loginURL, DefaultAdminLoginURL)
}

func (h *AuthHandler) successRedirectURL() string {
	if h == nil {
		return DefaultAdminRedirectURL
	}
	return normalizeRedirectTarget(h.redirectURL, DefaultAdminRedirectURL)
}

func normalizeRedirectTarget(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

type authExchangeCodeStore struct {
	mu      sync.Mutex
	entries map[string]authExchangeCodeEntry
	now     func() time.Time
	ttl     time.Duration
}

type authExchangeCodeEntry struct {
	sessionJWT string
	expiresAt  time.Time
	codeExpiry time.Time
}

func newAuthExchangeCodeStore(ttl time.Duration) *authExchangeCodeStore {
	return &authExchangeCodeStore{
		entries: make(map[string]authExchangeCodeEntry),
		now:     time.Now,
		ttl:     ttl,
	}
}

func (s *authExchangeCodeStore) Issue(sessionJWT string, expiresAt time.Time) (string, time.Duration, error) {
	if s == nil {
		return "", 0, errors.New("exchange code store is required")
	}

	now := s.now()
	codeExpiry := now.Add(s.ttl)
	if !expiresAt.IsZero() && expiresAt.Before(codeExpiry) {
		codeExpiry = expiresAt
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked(now)

	for attempt := 0; attempt < 5; attempt++ {
		code, err := generateExchangeCode()
		if err != nil {
			return "", 0, err
		}
		if _, exists := s.entries[code]; exists {
			continue
		}
		s.entries[code] = authExchangeCodeEntry{
			sessionJWT: sessionJWT,
			expiresAt:  expiresAt,
			codeExpiry: codeExpiry,
		}
		return code, time.Until(codeExpiry), nil
	}

	return "", 0, errors.New("unable to allocate exchange code")
}

func (s *authExchangeCodeStore) Consume(code string) (authExchangeCodeEntry, bool) {
	if s == nil {
		return authExchangeCodeEntry{}, false
	}

	now := s.now()
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return authExchangeCodeEntry{}, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanupExpiredLocked(now)

	entry, ok := s.entries[trimmed]
	if !ok {
		return authExchangeCodeEntry{}, false
	}
	delete(s.entries, trimmed)
	if !entry.codeExpiry.IsZero() && now.After(entry.codeExpiry) {
		return authExchangeCodeEntry{}, false
	}
	return entry, true
}

func (s *authExchangeCodeStore) cleanupExpiredLocked(now time.Time) {
	for code, entry := range s.entries {
		if !entry.codeExpiry.IsZero() && now.After(entry.codeExpiry) {
			delete(s.entries, code)
		}
	}
}

func generateExchangeCode() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("read exchange code entropy: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func readLegacyJWTFromExchangeRequest(r *http.Request) (string, error) {
	if r == nil {
		return "", errors.New("request is required")
	}

	if token, ok := bearerToken(strings.TrimSpace(r.Header.Get("Authorization"))); ok {
		return token, nil
	}

	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "application/json") {
		var payload struct {
			JWT string `json:"jwt"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&payload); err != nil {
			return "", err
		}
		token := strings.TrimSpace(payload.JWT)
		if token == "" {
			return "", errors.New("legacy jwt is required")
		}
		return token, nil
	}

	if err := r.ParseForm(); err != nil {
		return "", err
	}
	token := strings.TrimSpace(r.Form.Get("jwt"))
	if token == "" {
		return "", errors.New("legacy jwt is required")
	}
	return token, nil
}

func bearerToken(header string) (string, bool) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}
