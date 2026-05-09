package articleinspect

import (
	actionspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/actions"
	articlespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/articles"
	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
	lifecyclepkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/lifecycle"
	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	resultspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/results"
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
	"github.com/dovetaill/article-sentinel/pkg/config"
	"gorm.io/gorm"
)

// NewRoutes assembles the article inspection module dependencies for route registration.
func NewRoutes(db *gorm.DB, dispatcher outboxpkg.TaskDispatcher) Routes {
	if db == nil {
		return Routes{}
	}

	categoryRepo := rulespkg.NewCategoryRepository(db)
	keywordRepo := rulespkg.NewKeywordRepository(db)
	articleRepo := articlespkg.NewArticleRepository(db)

	return Routes{
		Categories: rulespkg.NewCategoryService(categoryRepo),
		Keywords:   rulespkg.NewKeywordService(keywordRepo),
		Tasks:      taskspkg.NewTaskService(db, keywordRepo, articleRepo),
		Results:    resultspkg.NewResultService(db),
		Actions:    actionspkg.NewActionService(db, actionspkg.NewActionRepository(db), nil),
		Lifecycle:  lifecyclepkg.NewLifecycleService(db),
		Logs:       auditpkg.NewLogService(db),
		Articles:   articlespkg.NewArticleService(articleRepo),
		Dispatcher: dispatcher,
		Outbox:     outboxpkg.NewTaskOutboxSettings(config.OutboxConfig{}),
	}
}
