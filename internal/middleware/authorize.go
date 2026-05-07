package middleware

import (
	"net/http"

	"github.com/dovetaill/go-auth-demo/internal/identity"
)

func RequireAuthenticated() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := identity.ActorFromContext(r.Context()); !ok {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(role string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := identity.ActorFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if !actor.HasRole(role) {
				writeAuthError(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdmin() Middleware {
	return RequireRole(string(identity.PrincipalAdmin))
}

func RequireMember() Middleware {
	return RequireRole(string(identity.PrincipalMember))
}
