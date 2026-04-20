package tasks

import (
	"encoding/json"
	"errors"
	"fmt"

	libasynq "github.com/hibiken/asynq"
)

const TypeArticleInspectRunTask = "articleinspect:run-task"

type ArticleInspectTaskPayload struct {
	TaskID        uint64 `json:"task_id"`
	OrgID         uint64 `json:"orgid"`
	TriggerSource string `json:"trigger_source"`
	OperatorID    uint64 `json:"operator_id"`
	OperatorName  string `json:"operator_name"`
}

func NewArticleInspectTask(payload ArticleInspectTaskPayload) (*libasynq.Task, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal article inspect payload: %w", err)
	}
	return libasynq.NewTask(TypeArticleInspectRunTask, body), nil
}

func DecodeArticleInspectTaskPayload(task *libasynq.Task) (ArticleInspectTaskPayload, error) {
	if task == nil {
		return ArticleInspectTaskPayload{}, errors.New("task is required")
	}

	var payload ArticleInspectTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return ArticleInspectTaskPayload{}, fmt.Errorf("decode article inspect payload: %w", err)
	}
	return payload, nil
}
