package articleinspect

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

type TaskDispatcher interface {
	DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error
}

type Routes struct {
	Categories *CategoryService
	Keywords   *KeywordService
	Tasks      *TaskService
	Results    *ResultService
	Actions    *ActionService
	Lifecycle  *LifecycleService
	Logs       *LogService
	Articles   *ArticleService
	Dispatcher TaskDispatcher
	Logger     *slog.Logger
}

func RegisterRoutes(api huma.API, routes Routes) {
	if api == nil {
		return
	}

	inspect := huma.NewGroup(api, "/api/v1/article-inspect")
	inspect.UseSimpleModifier(func(op *huma.Operation) {
		op.SkipValidateParams = true
	})

	if routes.Categories != nil {
		registerCategoryRoutes(inspect, routes.Categories)
	}
	if routes.Keywords != nil {
		registerKeywordRoutes(inspect, routes.Keywords)
	}
	if routes.Tasks != nil {
		registerTaskRoutes(inspect, routes.Tasks, routes.Dispatcher, routes.Logger)
	}
	if routes.Results != nil {
		registerResultRoutes(inspect, routes.Results)
	}
	if routes.Actions != nil {
		registerActionRoutes(inspect, routes.Actions)
	}
	if routes.Lifecycle != nil {
		registerLifecycleRoutes(inspect, routes.Lifecycle)
	}
	if routes.Logs != nil {
		registerLogRoutes(inspect, routes.Logs)
	}
	if routes.Articles != nil {
		registerArticleRoutes(inspect, routes.Articles)
	}
}
