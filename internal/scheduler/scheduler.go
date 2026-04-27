package scheduler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/robfig/cron/v3"
)

// New 构建一个最小 cron scheduler。
func New() *cron.Cron {
	return cron.New()
}

// RegisterJobs 按当前配置注册所有 scheduler jobs。
func RegisterJobs(c *cron.Cron, rt *bootstrap.Runtime, enqueuer Enqueuer) error {
	if c == nil {
		return errors.New("cron is required")
	}
	if rt == nil || rt.Config == nil {
		return errors.New("scheduler runtime config is required")
	}
	if !rt.Config.Scheduler.Enabled {
		// 默认关闭，避免本地或生产在未明确配置时误跑定时任务。
		return nil
	}
	if enqueuer == nil {
		return errors.New("scheduler enqueuer is required")
	}

	spec := strings.TrimSpace(rt.Config.Scheduler.Spec)
	if spec == "" {
		return errors.New("scheduler spec is required")
	}

	// 当前只保留一个 heartbeat 骨架 job，真实业务 cron 后续按同样模式扩展。
	if _, err := c.AddFunc(spec, NewRuntimeHeartbeatJob(rt.Logger, enqueuer)); err != nil {
		return fmt.Errorf("register runtime heartbeat job: %w", err)
	}
	return nil
}
