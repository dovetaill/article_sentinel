package articleinspect

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	actionspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/actions"
	articlespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/articles"
	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
	lifecyclepkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/lifecycle"
	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	resultspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/results"
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
)

type Routes struct {
	Categories *rulespkg.CategoryService
	Keywords   *rulespkg.KeywordService
	Tasks      *taskspkg.TaskService
	Results    *resultspkg.ResultService
	Actions    *actionspkg.ActionService
	Lifecycle  *lifecyclepkg.LifecycleService
	Logs       *auditpkg.LogService
	Articles   *articlespkg.ArticleService
	Dispatcher outboxpkg.TaskDispatcher
	Logger     *slog.Logger
	Outbox     outboxpkg.TaskOutboxSettings
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
		rulespkg.RegisterCategoryRoutes(inspect, routes.Categories)
	}
	if routes.Keywords != nil {
		rulespkg.RegisterKeywordRoutes(inspect, routes.Keywords)
	}
	if routes.Tasks != nil {
		taskspkg.RegisterTaskRoutes(inspect, routes.Tasks, routes.Dispatcher, routes.Logger, routes.Outbox)
	}
	if routes.Results != nil {
		resultspkg.RegisterResultRoutes(inspect, routes.Results)
	}
	if routes.Actions != nil {
		actionspkg.RegisterActionRoutes(inspect, routes.Actions)
	}
	if routes.Lifecycle != nil {
		lifecyclepkg.RegisterLifecycleRoutes(inspect, routes.Lifecycle)
	}
	if routes.Logs != nil {
		auditpkg.RegisterLogRoutes(inspect, routes.Logs)
	}
	if routes.Articles != nil {
		articlespkg.RegisterArticleRoutes(inspect, routes.Articles)
	}
}
