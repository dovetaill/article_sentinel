package queueasynq

import (
	"errors"
	"strings"

	"github.com/dovetaill/article-sentinel/internal/app/bootstrap"
	"github.com/dovetaill/article-sentinel/internal/queue/tasks"
	libasynq "github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Enqueuer 定义最小 enqueue 能力，方便测试替身注入。
type Enqueuer interface {
	Enqueue(task *libasynq.Task, opts ...libasynq.Option) (*libasynq.TaskInfo, error)
}

// NewClient 使用共享 Redis 连接构建 Asynq client。
func NewClient(rt *bootstrap.Runtime) (*libasynq.Client, error) {
	redisClient, err := runtimeRedis(rt)
	if err != nil {
		return nil, err
	}
	return libasynq.NewClientFromRedisClient(redisClient), nil
}

// EnqueueTask 将标准任务投递到指定队列。
func EnqueueTask(client Enqueuer, queueName string, payload tasks.Payload) (*libasynq.TaskInfo, error) {
	if client == nil {
		return nil, errors.New("enqueuer is required")
	}

	task, err := tasks.NewTask(payload)
	if err != nil {
		return nil, err
	}

	queueName = strings.TrimSpace(queueName)
	if queueName == "" {
		return client.Enqueue(task)
	}
	return client.Enqueue(task, libasynq.Queue(queueName))
}

// EnqueueArticleInspectTask 封装巡检任务投递，统一 payload 编码和队列选择逻辑。
func EnqueueArticleInspectTask(client Enqueuer, queueName string, payload tasks.ArticleInspectTaskPayload) (*libasynq.TaskInfo, error) {
	if client == nil {
		return nil, errors.New("enqueuer is required")
	}

	task, err := tasks.NewArticleInspectTask(payload)
	if err != nil {
		return nil, err
	}

	queueName = strings.TrimSpace(queueName)
	if queueName == "" {
		return client.Enqueue(task)
	}
	return client.Enqueue(task, libasynq.Queue(queueName))
}

func runtimeRedis(rt *bootstrap.Runtime) (*redis.Client, error) {
	if rt == nil || rt.Resources == nil || rt.Resources.Redis == nil {
		return nil, errors.New("worker runtime redis is required")
	}
	return rt.Resources.Redis, nil
}
