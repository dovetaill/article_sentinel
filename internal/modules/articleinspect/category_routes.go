package articleinspect

import (
	"github.com/danielgtaylor/huma/v2"
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
)

func registerCategoryRoutes(api huma.API, service *CategoryService) {
	rulespkg.RegisterCategoryRoutes(api, service)
}
