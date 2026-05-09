package articleinspect

import (
	articlespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/articles"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
	"gorm.io/gorm"
)

// NewRoutes assembles the article inspection module dependencies for route registration.
func NewRoutes(db *gorm.DB, dispatcher TaskDispatcher) Routes {
	if db == nil {
		return Routes{}
	}

	categoryRepo := NewCategoryRepository(db)
	keywordRepo := NewKeywordRepository(db)
	articleRepo := articlespkg.NewArticleRepository(db)

	return Routes{
		Categories: NewCategoryService(categoryRepo),
		Keywords:   NewKeywordService(keywordRepo),
		Tasks:      taskspkg.NewTaskService(db, keywordRepo, articleRepo),
		Results:    NewResultService(db),
		Actions:    NewActionService(db, NewActionRepository(db)),
		Lifecycle:  NewLifecycleService(db),
		Logs:       NewLogService(db),
		Articles:   NewArticleService(articleRepo),
		Dispatcher: dispatcher,
		Outbox:     defaultTaskOutboxSettings(),
	}
}
