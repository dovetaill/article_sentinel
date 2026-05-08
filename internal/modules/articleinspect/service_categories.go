package articleinspect

import rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"

var (
	ErrInvalidCategoryInput = rulespkg.ErrInvalidCategoryInput
	ErrCategoryNotFound     = rulespkg.ErrCategoryNotFound
)

type CategoryService = rulespkg.CategoryService

func NewCategoryService(repo *CategoryRepository) *CategoryService {
	return rulespkg.NewCategoryService(repo)
}
