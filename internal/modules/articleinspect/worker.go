package articleinspect

import (
	"context"

	workerpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/worker"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

type Worker struct {
	db          *gorm.DB
	scanner     Scanner
	articleRepo *ArticleRepository
	batchSize   int
}

// NewWorker 构建巡检异步执行器。
func NewWorker(db *gorm.DB) *Worker {
	return &Worker{
		db:          db,
		scanner:     NewKeywordScanner(),
		articleRepo: NewArticleRepository(db),
		batchSize:   100,
	}
}

// ExecuteTask 是一期巡检主链路：拉起任务、分页扫描文稿、落结果、回写任务状态。
func (w *Worker) ExecuteTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	executor := workerpkg.NewExecutorWithDeps(w.db, w.scanner, w.articleRepo, w.batchSize)
	return executor.ExecuteTask(ctx, payload)
}
