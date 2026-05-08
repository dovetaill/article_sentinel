package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
)

func (r *AuditRepository) CreateOperationLog(ctx context.Context, log *domainpkg.InspectionOperationLog) error {
	if r == nil || r.db == nil || log == nil {
		return ErrInvalidLogQuery
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditRepository) CreateFieldChangeLogs(ctx context.Context, logs []domainpkg.InspectionFieldChangeLog) error {
	if r == nil || r.db == nil {
		return ErrInvalidLogQuery
	}
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&logs).Error
}

func ResolveTaskID(requestTaskID, resultTaskID uint64) uint64 {
	if resultTaskID != 0 {
		return resultTaskID
	}
	return requestTaskID
}

func BuildSnapshot(payload any) string {
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func BuildOperationLogSummary(operationType, status, beforeState, afterState, reason string, taskID, articleID, resultID uint64) string {
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
	case domainpkg.ActionTypeBatchIgnore:
		return "ignore"
	case domainpkg.ActionTypeBatchProcess:
		return "process"
	case domainpkg.ActionTypeOffline:
		return "offline"
	case domainpkg.ActionTypeRectify:
		return "rectify"
	case domainpkg.ActionTypeRepublish:
		return "republish"
	default:
		return strings.TrimSpace(value)
	}
}

func operationStatusLabel(value string) string {
	switch strings.TrimSpace(value) {
	case domainpkg.ActionStatusSuccess:
		return "succeeded"
	case domainpkg.ActionStatusSkipped:
		return "skipped"
	case domainpkg.ActionStatusFailed:
		return "failed"
	default:
		return strings.TrimSpace(value)
	}
}
