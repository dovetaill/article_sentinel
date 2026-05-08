package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/pkg/database"
)

type readyOutput struct {
	Body response.Envelope
}

type dependencyState struct {
	Configured bool `json:"configured"`
	Healthy    bool `json:"healthy"`
}

const readyProbeTimeout = 2 * time.Second

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
		databaseHealthy := false
		redisHealthy := false

		if databaseConfigured {
			probeCtx, cancel := context.WithTimeout(ctx, readyProbeTimeout)
			databaseHealthy = database.PingDB(probeCtx, rt.Resources.DB) == nil
			cancel()
		}

		if redisConfigured {
			probeCtx, cancel := context.WithTimeout(ctx, readyProbeTimeout)
			redisHealthy = database.PingRedis(probeCtx, rt.Resources.Redis) == nil
			cancel()
		}

		deps := map[string]dependencyState{
			"database": {
				Configured: databaseConfigured,
				Healthy:    databaseHealthy,
			},
			"redis": {
				Configured: redisConfigured,
				Healthy:    redisHealthy,
			},
		}

		return &readyOutput{
			Body: response.OK("ready", map[string]any{
				"dependencies": deps,
			}),
		}, nil
	})
}
