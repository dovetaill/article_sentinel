package register

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/article-sentinel/internal/api/handlers"
	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/internal/identity"
	"github.com/dovetaill/article-sentinel/internal/middleware"
	"github.com/dovetaill/article-sentinel/pkg/config"
)

// NewRouter 构建基于 Huma 的 HTTP 路由。
func NewRouter(rt *bootstrap.Runtime) http.Handler {
	apiMux := http.NewServeMux()
	adminSessionManager := newAdminSessionManager(rt)
	handlers.RegisterAuthRoutes(apiMux, adminSessionManager, sessionConfig(rt))
	cfg := huma.DefaultConfig("article-sentinel API", "0.1.0")
	if rt != nil && rt.Config != nil {
		if rt.Config.App.Name != "" {
			cfg.Info.Title = rt.Config.App.Name
		}
		if rt.Config.Docs.Enabled {
			cfg.OpenAPIPath = normalizeOpenAPIPath(rt.Config.Docs.OpenAPIPath)
			if rt.Config.Docs.UIPath != "" {
				cfg.DocsPath = rt.Config.Docs.UIPath
			}
		} else {
			cfg.OpenAPIPath = ""
			cfg.DocsPath = ""
			cfg.SchemasPath = ""
		}
	}

	api := humago.New(apiMux, cfg)
	publicRoutes := huma.NewGroup(api)
	handlers.RegisterHealth(publicRoutes)
	handlers.RegisterReady(publicRoutes, rt)

	timeout := 15 * time.Second
	if rt != nil && rt.Config != nil && rt.Config.HTTP.RequestTimeoutSeconds > 0 {
		timeout = time.Duration(rt.Config.HTTP.RequestTimeoutSeconds) * time.Second
	}

	return middleware.Chain(
		apiMux,
		middleware.RequestID(),
		middleware.SessionContext(adminSessionManager),
		middleware.AccessLog(nilLogger(rt)),
		middleware.ProtectDocumentation(docsEnabled(rt), middleware.WithDocumentationPaths(documentationProtectionPaths(rt))),
		middleware.Recover(),
		middleware.Timeout(timeout),
	)
}

func normalizeOpenAPIPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/openapi"
	}
	if strings.HasSuffix(path, ".json") {
		return strings.TrimSuffix(path, ".json")
	}
	return path
}

func nilLogger(rt *bootstrap.Runtime) *slog.Logger {
	if rt == nil {
		return nil
	}
	return rt.Logger
}

func newAdminSessionManager(rt *bootstrap.Runtime) *identity.AdminSessionManager {
	if rt == nil || rt.Config == nil {
		return nil
	}
	return identity.NewAdminSessionManager(rt.Config.Auth.Session)
}

func sessionConfig(rt *bootstrap.Runtime) config.SessionConfig {
	if rt == nil || rt.Config == nil {
		return config.SessionConfig{}
	}
	return rt.Config.Auth.Session
}

func docsEnabled(rt *bootstrap.Runtime) bool {
	if rt == nil || rt.Config == nil {
		return true
	}
	return rt.Config.Docs.Enabled
}

func documentationProtectionPaths(rt *bootstrap.Runtime) middleware.DocumentationPaths {
	paths := middleware.DocumentationPaths{
		OpenAPIPath: "/openapi",
		DocsPath:    "/docs",
		SchemasPath: "/schemas",
	}
	if rt == nil || rt.Config == nil {
		return paths
	}

	paths.OpenAPIPath = normalizeOpenAPIPath(rt.Config.Docs.OpenAPIPath)
	if rt.Config.Docs.UIPath != "" {
		paths.DocsPath = rt.Config.Docs.UIPath
	}
	return paths
}
