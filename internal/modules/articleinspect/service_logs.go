package articleinspect

import (
	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
	"gorm.io/gorm"
)

var ErrInvalidLogQuery = auditpkg.ErrInvalidLogQuery

type OperationLogListInput = auditpkg.OperationLogListInput

type FieldChangeLogListInput = auditpkg.FieldChangeLogListInput

type OperationLogListResult = auditpkg.OperationLogListResult

type FieldChangeLogListResult = auditpkg.FieldChangeLogListResult

type LogService = auditpkg.LogService

func NewLogService(db *gorm.DB) *LogService {
	return auditpkg.NewLogService(db)
}
