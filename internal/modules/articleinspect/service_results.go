package articleinspect

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrInvalidResultQuery = errors.New("invalid result query")

type ResultService struct {
	repo *ResultRepository
}

type ResultListResult struct {
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
	Items    []InspectionResult `json:"items"`
}

type ResultDetail struct {
	Result          InspectionResult           `json:"result"`
	Hits            []InspectionResultHit      `json:"hits"`
	OperationLogs   []InspectionOperationLog   `json:"operation_logs"`
	FieldChangeLogs []InspectionFieldChangeLog `json:"field_change_logs"`
}

func NewResultService(db *gorm.DB) *ResultService {
	return &ResultService{repo: NewResultRepository(db)}
}

func (s *ResultService) List(ctx context.Context, input ResultListInput) (*ResultListResult, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 {
		return nil, ErrInvalidResultQuery
	}
	page, pageSize := normalizePage(input.Page, input.PageSize)
	items, total, err := s.repo.ListResults(ctx, input)
	if err != nil {
		return nil, err
	}
	return &ResultListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *ResultService) GetDetail(ctx context.Context, orgID, resultID uint64) (*ResultDetail, error) {
	if s == nil || s.repo == nil || orgID == 0 || resultID == 0 {
		return nil, ErrInvalidResultQuery
	}
	result, err := s.repo.GetResult(ctx, orgID, resultID)
	if err != nil {
		return nil, err
	}
	hits, err := s.repo.ListHits(ctx, orgID, resultID)
	if err != nil {
		return nil, err
	}
	operationLogs, _, err := s.repo.ListOperationLogs(ctx, OperationLogListInput{OrgID: orgID, ArticleID: result.ArticleID, TaskID: result.TaskID, Page: 1, PageSize: 20})
	if err != nil {
		return nil, err
	}
	fieldChangeLogs, _, err := s.repo.ListFieldChangeLogs(ctx, FieldChangeLogListInput{OrgID: orgID, ArticleID: result.ArticleID, Page: 1, PageSize: 20})
	if err != nil {
		return nil, err
	}
	return &ResultDetail{Result: *result, Hits: hits, OperationLogs: operationLogs, FieldChangeLogs: fieldChangeLogs}, nil
}
