package tasks

import (
	"encoding/json"
	"errors"
	"fmt"

	libasynq "github.com/hibiken/asynq"
)

// Payload 定义队列骨架任务载荷。
type Payload struct {
	Source string `json:"source"`
}

// DecodePayload 解析标准队列任务载荷。
func DecodePayload(task *libasynq.Task) (Payload, error) {
	if task == nil {
		return Payload{}, errors.New("task is required")
	}

	var payload Payload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return Payload{}, fmt.Errorf("decode payload: %w", err)
	}
	return payload, nil
}
