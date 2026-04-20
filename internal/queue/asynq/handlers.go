package queueasynq

import (
	"context"
	"fmt"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	articleinspectmodule "github.com/dovetaill/article-sentinel/internal/modules/articleinspect"
	"github.com/dovetaill/article-sentinel/internal/queue/tasks"
	libasynq "github.com/hibiken/asynq"
)

type articleInspectExecutor interface {
	ExecuteTask(ctx context.Context, payload tasks.ArticleInspectTaskPayload) error
}

var newArticleInspectExecutorFn = func(rt *bootstrap.Runtime) articleInspectExecutor {
	if rt == nil || rt.Resources == nil || rt.Resources.DB == nil {
		return nil
	}
	return articleinspectmodule.NewWorker(rt.Resources.DB)
}

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
	if executor := newArticleInspectExecutorFn(rt); executor != nil {
		mux.HandleFunc(tasks.TypeArticleInspectRunTask, func(ctx context.Context, task *libasynq.Task) error {
			payload, err := tasks.DecodeArticleInspectTaskPayload(task)
			if err != nil {
				return fmt.Errorf("decode %s payload: %w", tasks.TypeArticleInspectRunTask, err)
			}
			return executor.ExecuteTask(ctx, payload)
		})
	}
	return mux
}
