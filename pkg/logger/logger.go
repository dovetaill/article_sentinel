package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dovetaill/article-sentinel/pkg/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New 根据日志配置创建 JSON slog，并返回释放资源用的 cleanup。
func New(cfg config.LogConfig) (*slog.Logger, func() error, error) {
	writer, fileLogger, err := buildWriter(cfg)
	if err != nil {
		return nil, nil, err
	}

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: parseLevel(cfg.Level),
	})

	cleanup := func() error { return nil }
	if fileLogger != nil {
		stopRotate := func() error { return nil }
		if cfg.RotateDaily {
			stopRotate = startDailyRotation(timeNow, newStdTicker, fileLogger)
		}

		cleanup = func() error {
			return errors.Join(stopRotate(), fileLogger.Close())
		}
	}

	return slog.New(handler), cleanup, nil
}

func buildWriter(cfg config.LogConfig) (io.Writer, *lumberjack.Logger, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Output)) {
	case "", "stdout":
		return os.Stdout, nil, nil
	case "file":
		fileLogger, err := newFileLogger(cfg)
		if err != nil {
			return nil, nil, err
		}
		return fileLogger, fileLogger, nil
	case "both":
		fileLogger, err := newFileLogger(cfg)
		if err != nil {
			return nil, nil, err
		}
		return io.MultiWriter(os.Stdout, fileLogger), fileLogger, nil
	default:
		return nil, nil, fmt.Errorf("unsupported log output %q", cfg.Output)
	}
}

func newFileLogger(cfg config.LogConfig) (*lumberjack.Logger, error) {
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir %q: %w", cfg.Dir, err)
	}

	return &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, cfg.Filename),
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}, nil
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
