package bootstrap

import (
	"errors"
	"fmt"

	articleinspectmodule "github.com/dovetaill/article-sentinel/internal/modules/articleinspect"
	postmodule "github.com/dovetaill/article-sentinel/internal/modules/post"
)

type schemaMigrator interface {
	AutoMigrate(dst ...any) error
}

func init() {
	RegisterBusinessModels(
		postmodule.Post{},
		articleinspectmodule.InspectionKeyword{},
		articleinspectmodule.InspectionKeywordScope{},
		articleinspectmodule.InspectionTask{},
		articleinspectmodule.InspectionTaskKeyword{},
		articleinspectmodule.InspectionResult{},
		articleinspectmodule.InspectionResultHit{},
		articleinspectmodule.InspectionAction{},
		articleinspectmodule.InspectionOperationLog{},
		articleinspectmodule.InspectionFieldChangeLog{},
	)
}

func RegisterBusinessModels(models ...any) {
	for _, model := range models {
		if model == nil {
			continue
		}
		businessModels = append(businessModels, model)
	}
}

func AutoMigrateBusinessTables(migrator schemaMigrator) error {
	if migrator == nil {
		return errors.New("schema migrator is required")
	}
	if len(businessModels) == 0 {
		return nil
	}
	if err := migrator.AutoMigrate(businessModels...); err != nil {
		return fmt.Errorf("auto migrate business models: %w", err)
	}
	return nil
}
