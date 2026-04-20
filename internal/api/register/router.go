package register

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/article-sentinel/internal/api/handlers"
	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/internal/middleware"
	articleinspectmodule "github.com/dovetaill/article-sentinel/internal/modules/articleinspect"
	postmodule "github.com/dovetaill/article-sentinel/internal/modules/post"
	queueasynq "github.com/dovetaill/article-sentinel/internal/queue/asynq"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

// NewRouter 构建基于 Huma 的 HTTP 路由。
func NewRouter(rt *bootstrap.Runtime) http.Handler {
	apiMux := http.NewServeMux()
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
	articleinspectmodule.RegisterRoutes(publicRoutes, newArticleInspectRoutes(rt))

	timeout := 15 * time.Second
	if rt != nil && rt.Config != nil && rt.Config.HTTP.RequestTimeoutSeconds > 0 {
		timeout = time.Duration(rt.Config.HTTP.RequestTimeoutSeconds) * time.Second
	}

	return middleware.Chain(
		apiMux,
		middleware.RequestID(),
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

	db := rt.Resources.DB
	keywordRepo := articleinspectmodule.NewKeywordRepository(db)
	return articleinspectmodule.Routes{
		Keywords:   articleinspectmodule.NewKeywordService(keywordRepo),
		Tasks:      articleinspectmodule.NewTaskService(db, keywordRepo, articleinspectmodule.NewArticleRepository(db)),
		Results:    articleinspectmodule.NewResultService(db),
		Actions:    articleinspectmodule.NewActionService(db, articleinspectmodule.NewActionRepository(db)),
		Lifecycle:  articleinspectmodule.NewLifecycleService(db),
		Logs:       articleinspectmodule.NewLogService(db),
		Dispatcher: newArticleInspectDispatcher(rt),
	}
}

type articleInspectDispatcher struct {
	client    queueasynq.Enqueuer
	queueName string
}

func newArticleInspectDispatcher(rt *bootstrap.Runtime) articleinspectmodule.TaskDispatcher {
	client, err := queueasynq.NewClient(rt)
	if err != nil {
		return nil
	}
	if rt != nil {
		rt.RegisterCloser(client.Close)
	}
	queueName := ""
	if rt != nil && rt.Config != nil {
		queueName = rt.Config.Queue.Asynq.QueueName
	}
	return &articleInspectDispatcher{client: client, queueName: queueName}
}

func (d *articleInspectDispatcher) DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	_ = ctx
	if d == nil {
		return nil
	}
	_, err := queueasynq.EnqueueArticleInspectTask(d.client, d.queueName, payload)
	return err
}

func nilLogger(rt *bootstrap.Runtime) *slog.Logger {
	if rt == nil {
		return nil
	}
	return rt.Logger
}
