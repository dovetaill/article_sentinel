package articleinspect

import (
	"github.com/danielgtaylor/huma/v2"
	resultspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/results"
)

func registerResultRoutes(api huma.API, service *ResultService) {
	resultspkg.RegisterResultRoutes(api, service)
}
