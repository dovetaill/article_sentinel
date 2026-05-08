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
	articleinspectmodule "github.com/dovetaill/article-sentinel/internal/modules/articleinspect"
	postmodule "github.com/dovetaill/article-sentinel/internal/modules/post"
	queueasynq "github.com/dovetaill/article-sentinel/internal/queue/asynq"
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
		}
	}

	api := humago.New(apiMux, cfg)
	publicRoutes := huma.NewGroup(api)
	handlers.RegisterHealth(publicRoutes)
	handlers.RegisterReady(publicRoutes, rt)
	if postService := newPostService(rt); postService != nil {
		postmodule.RegisterRoutes(publicRoutes, postService)
	}
	// router 只负责把依赖组装进模块，不直接承载业务实现。
	articleinspectmodule.RegisterRoutes(publicRoutes, newArticleInspectRoutes(rt))

	timeout := 15 * time.Second
	if rt != nil && rt.Config != nil && rt.Config.HTTP.RequestTimeoutSeconds > 0 {
		timeout = time.Duration(rt.Config.HTTP.RequestTimeoutSeconds) * time.Second
	}

	return middleware.Chain(
		apiMux,
		middleware.RequestID(),
		middleware.RequestMetadata(trustedProxyCIDRs(rt)),
		middleware.SessionContext(adminSessionManager),
		middleware.Recover(),
		middleware.Timeout(timeout),
		middleware.AccessLog(nilLogger(rt)),
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

func newPostService(rt *bootstrap.Runtime) *postmodule.Service {
	if rt == nil || rt.Resources == nil || rt.Resources.DB == nil {
		return nil
	}
	repo := postmodule.NewRepository(rt.Resources.DB)
	return postmodule.NewService(repo)
}

func newArticleInspectRoutes(rt *bootstrap.Runtime) articleinspectmodule.Routes {
	if rt == nil || rt.Resources == nil || rt.Resources.DB == nil {
		return articleinspectmodule.Routes{}
	}

	// 任务创建接口只负责落库和投递，真正扫描放到 worker 中异步执行。
	routes := articleinspectmodule.NewRoutes(rt.Resources.DB, newArticleInspectDispatcher(rt))
	routes.Logger = nilLogger(rt)
	if rt.Config != nil {
		routes.Outbox = articleinspectmodule.NewTaskOutboxSettings(rt.Config.Queue.Outbox)
	}
	return routes
}

func newArticleInspectDispatcher(rt *bootstrap.Runtime) articleinspectmodule.TaskDispatcher {
	client, err := queueasynq.NewClient(rt)
	if err != nil {
		if logger := nilLogger(rt); logger != nil {
			logger.Error("article inspect dispatcher unavailable", "error", err)
		}
		return nil
	}
	if rt != nil {
		rt.RegisterCloser(client.Close)
	}
	queueName := ""
	if rt != nil && rt.Config != nil {
		queueName = rt.Config.Queue.Asynq.QueueName
	}
	return queueasynq.NewArticleInspectTaskDispatcher(client, queueName)
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

func trustedProxyCIDRs(rt *bootstrap.Runtime) []string {
	if rt == nil || rt.Config == nil {
		return nil
	}
	return rt.Config.HTTP.TrustedProxyCIDRs
}
