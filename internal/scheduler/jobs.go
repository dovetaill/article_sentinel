package scheduler

import (
	"context"
	"log/slog"

	queueasynq "github.com/dovetaill/article-sentinel/internal/queue/asynq"
	"github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

// Enqueuer 抽象定时任务的投递行为，便于测试替换。
type Enqueuer interface {
	EnqueueRuntimeHeartbeat(payload tasks.Payload) error
}

type ArticleInspectTaskOutboxRelay interface {
	RelayPendingArticleInspectTaskOutbox(ctx context.Context, limit int) (int, error)
}

type ArticleInspectTaskOutboxCleaner interface {
	CleanupArticleInspectTaskOutbox(ctx context.Context, limit int) (int, error)
}

type asynqEnqueuer struct {
	client    queueasynq.Enqueuer
	queueName string
}

// NewAsynqEnqueuer 用 Asynq client 适配调度器需要的 enqueue seam。
func NewAsynqEnqueuer(client queueasynq.Enqueuer, queueName string) Enqueuer {
	return &asynqEnqueuer{
		client:    client,
		queueName: queueName,
	}
}

func (e *asynqEnqueuer) EnqueueRuntimeHeartbeat(payload tasks.Payload) error {
	_, err := queueasynq.EnqueueTask(e.client, e.queueName, payload)
	return err
}

// NewRuntimeHeartbeatJob 生成只负责投递队列任务的 cron job。
func NewRuntimeHeartbeatJob(logger *slog.Logger, enqueuer Enqueuer) func() {
	return func() {
		if enqueuer == nil {
			return
		}

		// scheduler 只负责定时触发 enqueue，不在 cron 回调里做重业务逻辑。
		if err := enqueuer.EnqueueRuntimeHeartbeat(tasks.Payload{Source: "scheduler"}); err != nil && logger != nil {
			logger.Error("enqueue scheduled task", "type", tasks.TypeRuntimeHeartbeat, "error", err)
		}
	}
}

func NewArticleInspectTaskOutboxRelayJob(logger *slog.Logger, relay ArticleInspectTaskOutboxRelay, limit int) func() {
	return func() {
		if relay == nil {
			return
		}
		if _, err := relay.RelayPendingArticleInspectTaskOutbox(context.Background(), limit); err != nil && logger != nil {
			logger.Error("relay article inspect outbox", "error", err, "batch_size", limit)
		}
	}
}

func NewArticleInspectTaskOutboxCleanupJob(logger *slog.Logger, cleaner ArticleInspectTaskOutboxCleaner, limit int) func() {
	return func() {
		if cleaner == nil {
			return
		}
		if _, err := cleaner.CleanupArticleInspectTaskOutbox(context.Background(), limit); err != nil && logger != nil {
			logger.Error("cleanup article inspect outbox", "error", err, "batch_size", limit)
		}
	}
}
