package articleinspect

import (
	"github.com/danielgtaylor/huma/v2"
	articlespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/articles"
)

func registerArticleRoutes(api huma.API, service *ArticleService) {
	articlespkg.RegisterArticleRoutes(api, service)
}
