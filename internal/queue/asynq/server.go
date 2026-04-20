package queueasynq

import (
	"errors"
	"strings"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	libasynq "github.com/hibiken/asynq"
)

// NewServer 使用共享 Redis 连接构建 worker server。
func NewServer(rt *bootstrap.Runtime) (*libasynq.Server, error) {
	redisClient, err := runtimeRedis(rt)
	if err != nil {
		return nil, err
	}
	if rt == nil || rt.Config == nil {
		return nil, errors.New("worker runtime config is required")
	}

	queueName := strings.TrimSpace(rt.Config.Queue.Asynq.QueueName)
	if queueName == "" {
		queueName = "default"
	}
	concurrency := rt.Config.Queue.Asynq.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}

	return libasynq.NewServerFromRedisClient(redisClient, libasynq.Config{
		Concurrency: concurrency,
		Queues: map[string]int{
			queueName: 1,
		},
	}), nil
}
