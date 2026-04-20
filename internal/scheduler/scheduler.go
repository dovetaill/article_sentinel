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
		return nil
	}
	if enqueuer == nil {
		return errors.New("scheduler enqueuer is required")
	}

	spec := strings.TrimSpace(rt.Config.Scheduler.Spec)
	if spec == "" {
		return errors.New("scheduler spec is required")
	}

	if _, err := c.AddFunc(spec, NewRuntimeHeartbeatJob(rt.Logger, enqueuer)); err != nil {
		return fmt.Errorf("register runtime heartbeat job: %w", err)
	}
	return nil
}
