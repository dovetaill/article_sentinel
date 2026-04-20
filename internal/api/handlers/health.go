package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/article-sentinel/internal/api/response"
)

type healthOutput struct {
	Body response.Envelope
}

// RegisterHealth 注册 /healthz。
func RegisterHealth(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "healthz",
		Method:      http.MethodGet,
		Path:        "/healthz",
		Summary:     "health check",
	}, func(ctx context.Context, input *struct{}) (*healthOutput, error) {
		return &healthOutput{
			Body: response.OK("alive", map[string]any{"status": "up"}),
		}, nil
	})
}
