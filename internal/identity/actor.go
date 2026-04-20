package identity

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var (
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrActorDisabled = errors.New("actor disabled")
)

type Actor struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type actorContextKey struct{}

func NewActor(id uint, username, role, status string) Actor {
	return Actor{
		ID:       id,
		Username: strings.TrimSpace(username),
		Role:     strings.TrimSpace(role),
		Status:   strings.TrimSpace(status),
	}
}

func (a Actor) HasRole(role string) bool {
	role = strings.TrimSpace(role)
	return role != "" && strings.EqualFold(a.Role, role)
}

func ContextWithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorContextKey{}).(Actor)
	return actor, ok
}

func StatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrActorDisabled):
		return http.StatusUnauthorized, "user disabled"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
