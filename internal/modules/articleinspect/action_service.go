package articleinspect

import (
	actionspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/actions"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	"gorm.io/gorm"
)

var ErrInvalidActionInput = sharedpkg.ErrInvalidActionInput

type BatchActionInput = actionspkg.BatchActionInput

type BatchActionSummary = actionspkg.BatchActionSummary

type ActionService = actionspkg.ActionService

func NewActionService(db *gorm.DB, repo *ActionRepository) *ActionService {
	return actionspkg.NewActionService(db, repo, nil)
}
