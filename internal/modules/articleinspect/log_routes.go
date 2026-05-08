package articleinspect

import (
	"github.com/danielgtaylor/huma/v2"
	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
)

func registerLogRoutes(api huma.API, service *LogService) {
	auditpkg.RegisterLogRoutes(api, service)
}
