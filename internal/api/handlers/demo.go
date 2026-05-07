package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
)

type demoMeOutput struct {
	Body response.Envelope
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
			return nil, huma.Error401Unauthorized("unauthorized")
		}

		return &demoMeOutput{Body: response.OK("me", actor)}, nil
	})
}
