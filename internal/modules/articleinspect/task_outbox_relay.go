package articleinspect

import (
	"log/slog"

	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	"gorm.io/gorm"
)

type TaskOutboxRelay = outboxpkg.TaskOutboxRelay

func NewTaskOutboxRelay(db *gorm.DB, dispatcher TaskDispatcher, logger *slog.Logger) *TaskOutboxRelay {
	return outboxpkg.NewTaskOutboxRelay(db, dispatcher, logger)
}
