package articleinspect

import (
	"context"

	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

type Operator = sharedpkg.Operator

func ResolveOperator(ctx context.Context) Operator {
	return sharedpkg.ResolveOperator(ctx)
}

func enrichActionWithOperator(ctx context.Context, action *InspectionAction) {
	sharedpkg.EnrichActionWithOperator(ctx, action)
}

func enrichOperationLogWithOperator(ctx context.Context, log *InspectionOperationLog) {
	sharedpkg.EnrichOperationLogWithOperator(ctx, log)
}

func enrichFieldChangeLogsWithOperator(ctx context.Context, logs []InspectionFieldChangeLog) {
	sharedpkg.EnrichFieldChangeLogsWithOperator(ctx, logs)
}
