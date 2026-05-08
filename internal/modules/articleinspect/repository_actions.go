package articleinspect

import (
	actionspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/actions"
	"gorm.io/gorm"
)

type ActionRepository = actionspkg.ActionRepository

func NewActionRepository(db *gorm.DB) *ActionRepository {
	return actionspkg.NewActionRepository(db)
}
