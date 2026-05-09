package queueasynq

import (
	"context"
	"errors"

	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	libasynq "github.com/hibiken/asynq"
)

type articleInspectTaskDispatcher struct {
	client    Enqueuer
	queueName string
}

func NewArticleInspectTaskDispatcher(client Enqueuer, queueName string) outboxpkg.TaskDispatcher {
	if client == nil {
		return nil
	}
	return &articleInspectTaskDispatcher{client: client, queueName: queueName}
}

func (d *articleInspectTaskDispatcher) DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	_ = ctx
	_, err := EnqueueArticleInspectTask(d.client, d.queueName, payload)
	if errors.Is(err, libasynq.ErrTaskIDConflict) {
		return nil
	}
	return err
}
