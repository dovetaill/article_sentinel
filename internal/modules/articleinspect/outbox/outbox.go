package outbox

import (
	"context"
	"errors"

	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

var ErrTaskOutboxDispatcherUnavailable = errors.New("task outbox dispatcher unavailable")
var ErrInvalidTaskInput = errors.New("invalid task input")

type TaskDispatcher interface {
	DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error
}

type TaskOutboxDispatchReport struct {
	Scanned      int
	Claimed      int
	Dispatched   int
	Retried      int
	DeadLettered int
	Failed       int
}
