package articleinspect

import (
	"encoding/json"
	"strings"
	"time"

	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

func decodeTaskOutboxPayload(raw string) (queuetasks.ArticleInspectTaskPayload, error) {
	var payload queuetasks.ArticleInspectTaskPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return queuetasks.ArticleInspectTaskPayload{}, err
	}
	return payload, nil
}

func isDispatchableMessage(message InspectionTaskOutboxMessage, now time.Time) bool {
	switch strings.TrimSpace(message.Status) {
	case TaskOutboxStatusPending:
		return message.NextAttemptAt == nil || !message.NextAttemptAt.After(now)
	case TaskOutboxStatusClaimed:
		return message.ClaimUntil == nil || message.ClaimUntil.Before(now)
	default:
		return false
	}
}
