package articleinspect

import (
	"errors"
	"net/http"

	"github.com/dovetaill/article-sentinel/internal/identity"
	"gorm.io/gorm"
)

func articleInspectStatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, identity.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, ErrCategoryNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrKeywordNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrArticleNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrTaskNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrTaskDeleteNotAllowed):
		return http.StatusConflict, "task cannot be deleted"
	case errors.Is(err, ErrInvalidCategoryInput):
		return http.StatusBadRequest, "invalid category input"
	case errors.Is(err, ErrInvalidKeywordInput):
		return http.StatusBadRequest, "invalid keyword input"
	case errors.Is(err, ErrInvalidTaskInput):
		return http.StatusBadRequest, "invalid task input"
	case errors.Is(err, ErrInvalidResultQuery):
		return http.StatusBadRequest, "invalid result query"
	case errors.Is(err, ErrInvalidActionInput):
		return http.StatusBadRequest, "invalid action input"
	case errors.Is(err, ErrInvalidLogQuery):
		return http.StatusBadRequest, "invalid log query"
	case errors.Is(err, ErrInvalidArticleQuery):
		return http.StatusBadRequest, "invalid article query"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
