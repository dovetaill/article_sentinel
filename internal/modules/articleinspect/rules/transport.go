package rules

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
	case errors.Is(err, ErrCategoryNotFound), errors.Is(err, ErrKeywordNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrInvalidCategoryInput):
		return http.StatusBadRequest, "invalid category input"
	case errors.Is(err, ErrInvalidKeywordInput):
		return http.StatusBadRequest, "invalid keyword input"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func failureOKFromError(err error) (*sharedpkg.OKEnvelopeOutput, error) {
	status, message := statusFromError(err)
	return sharedpkg.FailureOKEnvelope(status, message), nil
}

func failureCreatedFromError(err error) (*sharedpkg.CreatedEnvelopeOutput, error) {
	status, message := statusFromError(err)
	return sharedpkg.FailureCreatedEnvelope(status, message), nil
}
