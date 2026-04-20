package articleinspect

import (
	"context"
	"strings"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

type Operator struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	RequestID string `json:"request_id"`
	SourceIP  string `json:"source_ip"`
}

func ResolveOperator(ctx context.Context) Operator {
	var operator Operator

	if actor, ok := identity.ActorFromContext(ctx); ok {
		operator.ID = uint64(actor.ID)
		operator.Name = strings.TrimSpace(actor.Username)
		operator.Role = strings.TrimSpace(actor.Role)
	}

	if principal, ok := identity.PrincipalFromContext(ctx); ok {
		if operator.ID == 0 {
			operator.ID = uint64(principal.UserID)
		}
		if operator.Name == "" {
			operator.Name = strings.TrimSpace(principal.Username)
		}
		if operator.Role == "" {
			operator.Role = strings.TrimSpace(principal.Role)
		}
	}

	if metadata, ok := identity.RequestMetadataFromContext(ctx); ok {
		operator.RequestID = strings.TrimSpace(metadata.RequestID)
		operator.SourceIP = strings.TrimSpace(metadata.SourceIP)
	}

	return operator
}

func enrichActionWithOperator(ctx context.Context, action *InspectionAction) {
	if action == nil {
		return
	}
	operator := ResolveOperator(ctx)
	if action.OperatorID == 0 {
		action.OperatorID = operator.ID
	}
	if strings.TrimSpace(action.OperatorName) == "" {
		action.OperatorName = operator.Name
	}
	action.RequestID = operator.RequestID
	action.SourceIP = operator.SourceIP
}

func enrichOperationLogWithOperator(ctx context.Context, log *InspectionOperationLog) {
	if log == nil {
		return
	}
	operator := ResolveOperator(ctx)
	if log.OperatorID == 0 {
		log.OperatorID = operator.ID
	}
	if strings.TrimSpace(log.OperatorName) == "" {
		log.OperatorName = operator.Name
	}
	log.RequestID = operator.RequestID
	log.SourceIP = operator.SourceIP
}

func enrichFieldChangeLogsWithOperator(ctx context.Context, logs []InspectionFieldChangeLog) {
	if len(logs) == 0 {
		return
	}
	operator := ResolveOperator(ctx)
	for index := range logs {
		if logs[index].OperatorID == 0 {
			logs[index].OperatorID = operator.ID
		}
		if strings.TrimSpace(logs[index].OperatorName) == "" {
			logs[index].OperatorName = operator.Name
		}
		logs[index].RequestID = operator.RequestID
		logs[index].SourceIP = operator.SourceIP
	}
}
