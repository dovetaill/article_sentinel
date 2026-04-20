package articleinspect

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrInvalidLogQuery = errors.New("invalid log query")

type LogService struct {
	repo *ResultRepository
}

type OperationLogListResult struct {
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
	Total    int64                    `json:"total"`
	Items    []InspectionOperationLog `json:"items"`
}

type FieldChangeLogListResult struct {
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
	Total    int64                      `json:"total"`
	Items    []InspectionFieldChangeLog `json:"items"`
}

func NewLogService(db *gorm.DB) *LogService {
	return &LogService{repo: NewResultRepository(db)}
}

func (s *LogService) ListOperationLogs(ctx context.Context, input OperationLogListInput) (*OperationLogListResult, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 {
		return nil, ErrInvalidLogQuery
	}
	page, pageSize := normalizePage(input.Page, input.PageSize)
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
	page, pageSize := normalizePage(input.Page, input.PageSize)
	items, total, err := s.repo.ListFieldChangeLogs(ctx, input)
	if err != nil {
		return nil, err
	}
	return &FieldChangeLogListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}
