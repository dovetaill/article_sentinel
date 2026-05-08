package audit

import (
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
)

type OperationLogListInput struct {
	OrgID        uint64
	ArticleID    uint64
	TaskID       uint64
	OperatorName string
	StartAt      *time.Time
	EndAt        *time.Time
	Page         int
	PageSize     int
}

type FieldChangeLogListInput struct {
	OrgID     uint64
	ArticleID uint64
	FieldName string
	StartAt   *time.Time
	EndAt     *time.Time
	Page      int
	PageSize  int
}

type OperationLogListResult struct {
	Page     int                                `json:"page"`
	PageSize int                                `json:"page_size"`
	Total    int64                              `json:"total"`
	Items    []domainpkg.InspectionOperationLog `json:"items"`
}

type FieldChangeLogListResult struct {
	Page     int                                  `json:"page"`
	PageSize int                                  `json:"page_size"`
	Total    int64                                `json:"total"`
	Items    []domainpkg.InspectionFieldChangeLog `json:"items"`
}
