package articleinspect

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
)

func registerTaskRoutes(api huma.API, service *TaskService, dispatcher TaskDispatcher, logger *slog.Logger, outboxSettings TaskOutboxSettings) {
	taskspkg.RegisterTaskRoutes(api, service, dispatcher, logger, outboxpkg.TaskOutboxSettings(outboxSettings))
}
