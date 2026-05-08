package audit

import (
	"context"
	"errors"

	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	"gorm.io/gorm"
)

var ErrInvalidLogQuery = errors.New("invalid log query")

type LogService struct {
	repo *AuditRepository
}

func NewLogService(db *gorm.DB) *LogService {
	return &LogService{repo: NewAuditRepository(db)}
}

func (s *LogService) ListOperationLogs(ctx context.Context, input OperationLogListInput) (*OperationLogListResult, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 {
		return nil, ErrInvalidLogQuery
	}
	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	items, total, err := s.repo.ListOperationLogs(ctx, input)
	if err != nil {
		return nil, err
	}
	return &OperationLogListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *LogService) ListFieldChangeLogs(ctx context.Context, input FieldChangeLogListInput) (*FieldChangeLogListResult, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 {
		return nil, ErrInvalidLogQuery
	}
	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	items, total, err := s.repo.ListFieldChangeLogs(ctx, input)
	if err != nil {
		return nil, err
	}
	return &FieldChangeLogListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}
