package tasks

import (
	"encoding/json"
	"fmt"

	libasynq "github.com/hibiken/asynq"
)

const TypeRuntimeHeartbeat = "runtime:heartbeat"

// NewTask 构建队列骨架使用的标准任务。
func NewTask(payload Payload) (*libasynq.Task, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal task payload: %w", err)
	}
	return libasynq.NewTask(TypeRuntimeHeartbeat, body), nil
}
