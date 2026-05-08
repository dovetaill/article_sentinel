package lifecycle

import (
	"errors"
	"net/http"

	"github.com/dovetaill/article-sentinel/internal/identity"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	"gorm.io/gorm"
)

func statusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, identity.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, sharedpkg.ErrInvalidActionInput):
		return http.StatusBadRequest, "invalid action input"
	case errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound, "resource not found"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func failureOKFromError(err error) (*sharedpkg.OKEnvelopeOutput, error) {
	status, message := statusFromError(err)
	return sharedpkg.FailureOKEnvelope(status, message), nil
}
