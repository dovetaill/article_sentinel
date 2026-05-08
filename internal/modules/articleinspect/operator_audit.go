package articleinspect

import "context"

func auditOperatorFromContext(ctx context.Context) (uint64, string) {
	operator := ResolveOperator(ctx)
	return operator.ID, operator.Name
}
