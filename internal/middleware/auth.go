package middleware

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/dovetaill/go-auth-demo/internal/api/response"
	"github.com/dovetaill/go-auth-demo/internal/identity"
)

type authenticator interface {
	Authenticate(ctx context.Context, token string) (*identity.Actor, error)
}

type HeaderAuthConfig struct {
	SourceHeader   string
	SourceValue    string
	UserIDHeader   string
	UsernameHeader string
	RoleHeader     string
	StatusHeader   string
}

type authOptions struct {
	trustedHeader *HeaderAuthConfig
	devHeader     *HeaderAuthConfig
}

type AuthOption func(*authOptions)

func WithTrustedHeader(config HeaderAuthConfig) AuthOption {
	normalized := normalizeHeaderAuthConfig(config)
	return func(options *authOptions) {
		options.trustedHeader = normalized
	}
}

func WithDevHeader(config HeaderAuthConfig) AuthOption {
	normalized := normalizeHeaderAuthConfig(config)
	return func(options *authOptions) {
		options.devHeader = normalized
	}
}

func Authenticate(authenticator authenticator, options ...AuthOption) Middleware {
	config := authOptions{}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := identity.ContextWithRequestMetadata(r.Context(), identity.RequestMetadata{
				SourceIP: clientIPFromRequest(r),
			})

			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				if actor, ok := actorFromHeaders(r, config.trustedHeader); ok {
					next.ServeHTTP(w, r.WithContext(contextWithActor(ctx, actor)))
					return
				}
				if actor, ok := actorFromHeaders(r, config.devHeader); ok {
					next.ServeHTTP(w, r.WithContext(contextWithActor(ctx, actor)))
					return
				}
				next.ServeHTTP(w, r.WithContext(ctx))
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

			next.ServeHTTP(w, r.WithContext(contextWithActor(ctx, *actor)))
		})
	}
}

func normalizeHeaderAuthConfig(config HeaderAuthConfig) *HeaderAuthConfig {
	config.SourceHeader = strings.TrimSpace(config.SourceHeader)
	config.SourceValue = strings.TrimSpace(config.SourceValue)
	config.UserIDHeader = strings.TrimSpace(config.UserIDHeader)
	config.UsernameHeader = strings.TrimSpace(config.UsernameHeader)
	config.RoleHeader = strings.TrimSpace(config.RoleHeader)
	config.StatusHeader = strings.TrimSpace(config.StatusHeader)
	if config.UserIDHeader == "" || config.UsernameHeader == "" {
		return nil
	}
	return &config
}

func actorFromHeaders(r *http.Request, config *HeaderAuthConfig) (identity.Actor, bool) {
	if r == nil || config == nil {
		return identity.Actor{}, false
	}
	if config.SourceHeader != "" && !strings.EqualFold(strings.TrimSpace(r.Header.Get(config.SourceHeader)), config.SourceValue) {
		return identity.Actor{}, false
	}

	rawID := strings.TrimSpace(r.Header.Get(config.UserIDHeader))
	username := strings.TrimSpace(r.Header.Get(config.UsernameHeader))
	if rawID == "" || username == "" {
		return identity.Actor{}, false
	}

	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return identity.Actor{}, false
	}

	status := strings.TrimSpace(r.Header.Get(config.StatusHeader))
	if status == "" {
		status = "active"
	}

	return identity.NewActor(uint(id), username, strings.TrimSpace(r.Header.Get(config.RoleHeader)), status), true
}

func contextWithActor(ctx context.Context, actor identity.Actor) context.Context {
	ctx = identity.ContextWithActor(ctx, actor)
	return identity.ContextWithPrincipal(ctx, identity.PrincipalFromActor(actor))
}

func clientIPFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	remoteAddr := strings.TrimSpace(r.RemoteAddr)
	if remoteAddr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
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
