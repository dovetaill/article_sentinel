package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
)

type demoMeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

type demoActor struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// RegisterDemoRoutes registers minimal protected demo endpoints.
func RegisterDemoRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "demo-me",
		Method:      http.MethodGet,
		Path:        "/api/v1/demo/me",
		Summary:     "current demo actor",
	}, func(ctx context.Context, input *struct{}) (*demoMeOutput, error) {
		actor, ok := identity.ActorFromContext(ctx)
		if !ok {
			return &demoMeOutput{
				Status: http.StatusUnauthorized,
				Body:   response.Fail(http.StatusUnauthorized, "unauthorized"),
			}, nil
		}

		return &demoMeOutput{
			Status: http.StatusOK,
			Body: response.OK("me", demoActor{
				ID:       actor.ID,
				Username: actor.Username,
				Role:     actor.Role,
				Status:   actor.Status,
			}),
		}, nil
	})
}
