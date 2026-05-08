package outbox

import (
	"encoding/json"
	"strings"
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

func decodeTaskOutboxPayload(raw string) (queuetasks.ArticleInspectTaskPayload, error) {
	var payload queuetasks.ArticleInspectTaskPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return queuetasks.ArticleInspectTaskPayload{}, err
	}
	return payload, nil
}

func isDispatchableMessage(message domainpkg.InspectionTaskOutboxMessage, now time.Time) bool {
	switch strings.TrimSpace(message.Status) {
	case domainpkg.TaskOutboxStatusPending:
		return message.NextAttemptAt == nil || !message.NextAttemptAt.After(now)
	case domainpkg.TaskOutboxStatusClaimed:
		return message.ClaimUntil == nil || message.ClaimUntil.Before(now)
	default:
		return false
	}
}
