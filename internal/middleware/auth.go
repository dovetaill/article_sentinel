package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
)

type authenticator interface {
	Authenticate(ctx context.Context, token string) (*identity.Actor, error)
}

func Authenticate(authenticator authenticator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			token, ok := bearerToken(header)
			if !ok || authenticator == nil {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			actor, err := authenticator.Authenticate(r.Context(), token)
			if err != nil {
				status, message := identity.StatusFromError(err)
				writeAuthError(w, status, message)
				return
			}

			ctx := identity.ContextWithActor(r.Context(), *actor)
			ctx = identity.ContextWithPrincipal(ctx, identity.PrincipalFromActor(*actor))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response.Fail(status, message))
}
