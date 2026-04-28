package articleinspect

import (
	"encoding/json"
	"fmt"
	"strings"
)

func resolveAuditTaskID(requestTaskID, resultTaskID uint64) uint64 {
	if resultTaskID != 0 {
		return resultTaskID
	}
	return requestTaskID
}

func buildAuditSnapshot(payload any) string {
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func buildOperationLogSummary(operationType, status, beforeState, afterState, reason string, taskID, articleID, resultID uint64) string {
	summary := strings.TrimSpace(strings.Join([]string{
		operationTypeLabel(operationType),
		operationStatusLabel(status),
	}, " "))

	scope := make([]string, 0, 3)
	if taskID != 0 {
		scope = append(scope, fmt.Sprintf("task #%d", taskID))
	}
	if articleID != 0 {
		scope = append(scope, fmt.Sprintf("article #%d", articleID))
	}
	if resultID != 0 {
		scope = append(scope, fmt.Sprintf("result #%d", resultID))
	}
	if len(scope) > 0 {
		summary += " for " + strings.Join(scope, ", ")
	}

	before := strings.TrimSpace(beforeState)
	after := strings.TrimSpace(afterState)
	if before != "" || after != "" {
		if before == "" {
			before = "-"
		}
		if after == "" {
			after = "-"
		}
		summary += fmt.Sprintf(" (%s -> %s)", before, after)
	}

	if reason := strings.TrimSpace(reason); reason != "" {
		summary += ": " + reason
	}

	return strings.TrimSpace(summary)
}

func operationTypeLabel(value string) string {
	switch strings.TrimSpace(value) {
	case ActionTypeBatchIgnore:
		return "ignore"
	case ActionTypeBatchProcess:
		return "process"
	case ActionTypeOffline:
		return "offline"
	case ActionTypeRectify:
		return "rectify"
	case ActionTypeRepublish:
		return "republish"
	default:
		return strings.TrimSpace(value)
	}
}

func operationStatusLabel(value string) string {
	switch strings.TrimSpace(value) {
	case ActionStatusSuccess:
		return "succeeded"
	case ActionStatusSkipped:
		return "skipped"
	case ActionStatusFailed:
		return "failed"
	default:
		return strings.TrimSpace(value)
	}
}
