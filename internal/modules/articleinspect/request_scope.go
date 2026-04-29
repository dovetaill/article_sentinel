package articleinspect

import (
	"context"
	"strings"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

func currentOrgID(ctx context.Context) (uint64, error) {
	session, ok := identity.AdminSessionFromContext(ctx)
	if !ok || session.OrgID == 0 {
		return 0, identity.ErrUnauthorized
	}
	return session.OrgID, nil
}

func validateOptionalOrgID(param uint64Param) error {
	if strings.TrimSpace(param.Raw) == "" {
		return nil
	}
	_, err := param.Parse()
	return err
}
