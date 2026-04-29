package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

func SessionContext(manager *identity.AdminSessionManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if manager == nil {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(manager.CookieName())
			if err != nil {
				if !errors.Is(err, http.ErrNoCookie) {
					clearSessionCookie(w, manager)
				}
				next.ServeHTTP(w, r)
				return
			}

			session, err := manager.ParseSessionJWT(cookie.Value)
			if err != nil {
				clearSessionCookie(w, manager)
				next.ServeHTTP(w, r)
				return
			}

			actor := session.Actor()
			ctx := identity.ContextWithAdminSession(r.Context(), session)
			ctx = identity.ContextWithActor(ctx, actor)
			ctx = identity.ContextWithPrincipal(ctx, identity.PrincipalFromActor(actor))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func clearSessionCookie(w http.ResponseWriter, manager *identity.AdminSessionManager) {
	if w == nil || manager == nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     manager.CookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   manager.SecureCookie(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}
