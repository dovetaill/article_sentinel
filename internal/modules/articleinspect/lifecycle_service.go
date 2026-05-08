package articleinspect

import (
	lifecyclepkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/lifecycle"
	"gorm.io/gorm"
)

type OfflineArticleInput = lifecyclepkg.OfflineArticleInput

type UpdateArticleFieldsInput = lifecyclepkg.UpdateArticleFieldsInput

type RepublishArticleInput = lifecyclepkg.RepublishArticleInput

type LifecycleActionResult = lifecyclepkg.LifecycleActionResult

type LifecycleService = lifecyclepkg.LifecycleService

func NewLifecycleService(db *gorm.DB) *LifecycleService {
	return lifecyclepkg.NewLifecycleService(db)
}
