package articleinspect

import (
	"context"

	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

func currentOrgID(ctx context.Context) (uint64, error) {
	return sharedpkg.CurrentOrgID(ctx)
}

func validateOptionalOrgID(param uint64Param) error {
	return sharedpkg.ValidateOptionalOrgID(param)
}
