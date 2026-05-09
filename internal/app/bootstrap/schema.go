package bootstrap

import (
	"errors"
	"fmt"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	postmodule "github.com/dovetaill/article-sentinel/internal/modules/post"
)

type schemaMigrator interface {
	AutoMigrate(dst ...any) error
}

func init() {
	// 新增业务表后，记得把 model 注册到这里，migrate 入口才能自动同步结构。
	RegisterBusinessModels(
		postmodule.Post{},
		domainpkg.InspectionKeyword{},
		domainpkg.InspectionKeywordScope{},
		domainpkg.InspectionTask{},
		domainpkg.InspectionTaskKeyword{},
		domainpkg.InspectionTaskOutboxMessage{},
		domainpkg.InspectionResult{},
		domainpkg.InspectionResultHit{},
		domainpkg.InspectionAction{},
		domainpkg.InspectionOperationLog{},
		domainpkg.InspectionFieldChangeLog{},
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
