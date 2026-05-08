package articleinspect

import (
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
	"gorm.io/gorm"
)

type CategoryRepository = rulespkg.CategoryRepository

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return rulespkg.NewCategoryRepository(db)
}
