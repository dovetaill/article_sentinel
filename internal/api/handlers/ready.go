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

type dependencyState struct {
	Configured bool `json:"configured"`
	Healthy    bool `json:"healthy"`
}

// RegisterReady 注册 /readyz。
func RegisterReady(api huma.API, rt *bootstrap.Runtime) {
	huma.Register(api, huma.Operation{
		OperationID: "readyz",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Summary:     "readiness check",
	}, func(ctx context.Context, input *struct{}) (*readyOutput, error) {
		databaseConfigured := rt != nil && rt.Resources != nil && rt.Resources.DB != nil
		redisConfigured := rt != nil && rt.Resources != nil && rt.Resources.Redis != nil

		deps := map[string]dependencyState{
			"database": {
				Configured: databaseConfigured,
				Healthy:    databaseConfigured,
			},
			"redis": {
				Configured: redisConfigured,
				Healthy:    redisConfigured,
			},
		}

		return &readyOutput{
			Body: response.OK("ready", map[string]any{
				"dependencies": deps,
			}),
		}, nil
	})
}
