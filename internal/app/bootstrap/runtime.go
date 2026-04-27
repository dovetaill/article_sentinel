package bootstrap

import (
	"fmt"
	"log/slog"

	"github.com/dovetaill/article-sentinel/internal/app/lifecycle"
	"github.com/dovetaill/article-sentinel/pkg/config"
	"github.com/dovetaill/article-sentinel/pkg/database"
	"github.com/dovetaill/article-sentinel/pkg/logger"
)

var (
	loadConfigFn                = config.Load
	newLoggerFn                 = logger.New
	bootstrapDatabaseFn         = database.Bootstrap
	businessModels              []any
	applySQLMigrationsFn        = ApplySQLMigrations
	autoMigrateBusinessTablesFn = AutoMigrateBusinessTables
)

// Runtime 承载 server/worker/scheduler 共享资源。
type Runtime struct {
	Config    *config.Config
	Logger    *slog.Logger
	Resources *database.Resources
	closers   []lifecycle.Closer
}

// RegisterCloser 追加一个需要在退出时执行的关闭动作。
func (r *Runtime) RegisterCloser(closer lifecycle.Closer) {
	if r == nil || closer == nil {
		return
	}
	r.closers = append(r.closers, closer)
}

// Shutdown 统一释放 Runtime 中注册的资源。
func (r *Runtime) Shutdown() error {
	if r == nil {
		return nil
	}
	return lifecycle.Shutdown(r.closers...)
}

func buildRuntime(configPath string) (*Runtime, error) {
	cfg, err := loadConfigFn(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// 启动顺序保持固定：先配置，再日志，再数据库/Redis 等共享资源。
	log, logCloser, err := newLoggerFn(cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("bootstrap logger: %w", err)
	}

	resources, err := bootstrapDatabaseFn(cfg)
	if err != nil {
		if logCloser != nil {
			_ = logCloser()
		}
		return nil, fmt.Errorf("bootstrap database resources: %w", err)
	}

	rt := &Runtime{
		Config:    cfg,
		Logger:    log,
		Resources: resources,
	}

	// 统一在 Runtime 中登记关闭动作，server/worker/scheduler 退出时走同一套收口。
	if logCloser != nil {
		rt.RegisterCloser(logCloser)
	}
	if resources != nil {
		rt.RegisterCloser(resources.Close)
	}

	return rt, nil
}
