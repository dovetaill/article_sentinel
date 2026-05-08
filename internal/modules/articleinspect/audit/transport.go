package audit

import (
	"errors"
	"net/http"

	"github.com/dovetaill/article-sentinel/internal/identity"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

func statusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, identity.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, ErrInvalidLogQuery):
		return http.StatusBadRequest, "invalid log query"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func failureOKFromError(err error) (*sharedpkg.OKEnvelopeOutput, error) {
	status, message := statusFromError(err)
	return sharedpkg.FailureOKEnvelope(status, message), nil
}
