package outbox

import (
	"os"
	"strings"
	"time"

	"github.com/dovetaill/article-sentinel/pkg/config"
)

type TaskOutboxSettings struct {
	LeaseDuration       time.Duration
	MaxAttempts         int
	DispatchedRetention time.Duration
	DeadLetterRetention time.Duration
}

func NewTaskOutboxSettings(cfg config.OutboxConfig) TaskOutboxSettings {
	settings := defaultTaskOutboxSettings()
	if cfg.LeaseDurationSeconds > 0 {
		settings.LeaseDuration = time.Duration(cfg.LeaseDurationSeconds) * time.Second
	}
	if cfg.MaxAttempts > 0 {
		settings.MaxAttempts = cfg.MaxAttempts
	}
	if cfg.DispatchedRetentionHours > 0 {
		settings.DispatchedRetention = time.Duration(cfg.DispatchedRetentionHours) * time.Hour
	}
	if cfg.DeadLetterRetentionHours > 0 {
		settings.DeadLetterRetention = time.Duration(cfg.DeadLetterRetentionHours) * time.Hour
	}
	return settings
}

func defaultTaskOutboxSettings() TaskOutboxSettings {
	return TaskOutboxSettings{
		LeaseDuration:       30 * time.Second,
		MaxAttempts:         10,
		DispatchedRetention: 7 * 24 * time.Hour,
		DeadLetterRetention: 30 * 24 * time.Hour,
	}
}

func defaultTaskOutboxClaimOwner() string {
	host, err := os.Hostname()
	if err != nil || strings.TrimSpace(host) == "" {
		return "articleinspect-outbox"
	}
	return "articleinspect-outbox@" + strings.TrimSpace(host)
}
