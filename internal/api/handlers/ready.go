package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
)

type readyOutput struct {
	Body response.Envelope
}

// RegisterReady 注册 /readyz。
func RegisterReady(api huma.API, _ *bootstrap.Runtime) {
	huma.Register(api, huma.Operation{
		OperationID: "readyz",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Summary:     "readiness check",
	}, func(ctx context.Context, input *struct{}) (*readyOutput, error) {
		return &readyOutput{
			Body: response.OK("ready", map[string]any{
				"status": "ready",
			}),
		}, nil
	})
}
