package queueasynq

import (
	"context"
	"fmt"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/internal/queue/tasks"
	libasynq "github.com/hibiken/asynq"
)

// RegisterHandlers 注册当前 worker 支持的任务处理函数。
func RegisterHandlers(rt *bootstrap.Runtime) *libasynq.ServeMux {
	mux := libasynq.NewServeMux()
	mux.HandleFunc(tasks.TypeRuntimeHeartbeat, func(ctx context.Context, task *libasynq.Task) error {
		_ = ctx
		payload, err := tasks.DecodePayload(task)
		if err != nil {
			return fmt.Errorf("decode %s payload: %w", tasks.TypeRuntimeHeartbeat, err)
		}
		if rt != nil && rt.Logger != nil {
			rt.Logger.Info("processed queue task", "type", task.Type(), "source", payload.Source)
		}
		return nil
	})
	return mux
}
