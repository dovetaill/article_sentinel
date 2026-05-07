package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dovetaill/go-auth-demo/internal/api/response"
	"github.com/dovetaill/go-auth-demo/internal/identity"
	"github.com/dovetaill/go-auth-demo/pkg/config"
)

const (
	DefaultAdminLoginURL    = "/login"
	DefaultAdminRedirectURL = "/"
)

type AuthHandler struct {
	manager     *identity.AdminSessionManager
	loginURL    string
	redirectURL string
}

func RegisterAuthRoutes(mux *http.ServeMux, manager *identity.AdminSessionManager, sessionCfg config.SessionConfig) {
	if mux == nil {
		return
	}

	handler := &AuthHandler{
		manager:     manager,
		loginURL:    normalizeRedirectTarget(sessionCfg.LoginURL, DefaultAdminLoginURL),
		redirectURL: normalizeRedirectTarget(sessionCfg.RedirectURL, DefaultAdminRedirectURL),
	}
	mux.Handle("/auth/login", http.HandlerFunc(handler.Login))
	mux.Handle("/api/v1/auth/session", http.HandlerFunc(handler.Session))
	mux.Handle("/api/v1/auth/logout", http.HandlerFunc(handler.Logout))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if h == nil || h.manager == nil {
		clearAdminSessionCookie(w, nil)
		http.Redirect(w, r, h.loginRedirectURL(), http.StatusFound)
		return
	}

	legacyJWT := strings.TrimSpace(r.URL.Query().Get("jwt"))
	if legacyJWT == "" {
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

	http.SetCookie(w, &http.Cookie{
		Name:     h.manager.CookieName(),
		Value:    sessionJWT,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.manager.SecureCookie(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.manager.TTL().Seconds()),
		Expires:  expiresAt,
	})
	http.Redirect(w, r, h.successRedirectURL(), http.StatusFound)
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
