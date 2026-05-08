package articleinspect

import (
	"github.com/danielgtaylor/huma/v2"
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
)

func registerKeywordRoutes(api huma.API, service *KeywordService) {
	rulespkg.RegisterKeywordRoutes(api, service)
}
