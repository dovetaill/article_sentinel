package articleinspect

import (
	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	"github.com/dovetaill/article-sentinel/pkg/config"
)

type TaskOutboxSettings = outboxpkg.TaskOutboxSettings

func NewTaskOutboxSettings(cfg config.OutboxConfig) TaskOutboxSettings {
	return outboxpkg.NewTaskOutboxSettings(cfg)
}

func defaultTaskOutboxSettings() TaskOutboxSettings {
	return outboxpkg.NewTaskOutboxSettings(config.OutboxConfig{})
}
