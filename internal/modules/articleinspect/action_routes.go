package articleinspect

import (
	"github.com/danielgtaylor/huma/v2"
	actionspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/actions"
)

func registerActionRoutes(api huma.API, service *ActionService) {
	actionspkg.RegisterActionRoutes(api, service)
}
