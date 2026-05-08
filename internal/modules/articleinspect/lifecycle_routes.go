package articleinspect

import (
	"github.com/danielgtaylor/huma/v2"
	lifecyclepkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/lifecycle"
)

func registerLifecycleRoutes(api huma.API, service *LifecycleService) {
	lifecyclepkg.RegisterLifecycleRoutes(api, service)
}
