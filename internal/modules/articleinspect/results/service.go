package results

import (
	"context"
	"errors"

	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	"gorm.io/gorm"
)

var ErrInvalidResultQuery = errors.New("invalid result query")

type auditReader interface {
	ListOperationLogs(ctx context.Context, input auditpkg.OperationLogListInput) ([]domainpkg.InspectionOperationLog, int64, error)
	ListFieldChangeLogs(ctx context.Context, input auditpkg.FieldChangeLogListInput) ([]domainpkg.InspectionFieldChangeLog, int64, error)
}

type ResultService struct {
	repo  *ResultRepository
	audit auditReader
}

func NewResultService(db *gorm.DB) *ResultService {
	return &ResultService{repo: NewResultRepository(db), audit: auditpkg.NewAuditRepository(db)}
}

func (s *ResultService) List(ctx context.Context, input ResultListInput) (*ResultListResult, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 {
		return nil, ErrInvalidResultQuery
	}
	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	items, total, err := s.repo.ListResults(ctx, input)
	if err != nil {
		return nil, err
	}
	return &ResultListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *ResultService) GetDetail(ctx context.Context, orgID, resultID uint64) (*ResultDetail, error) {
	if s == nil || s.repo == nil || s.audit == nil || orgID == 0 || resultID == 0 {
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
	operationLogs, _, err := s.audit.ListOperationLogs(ctx, auditpkg.OperationLogListInput{OrgID: orgID, ArticleID: result.ArticleID, TaskID: result.TaskID, Page: 1, PageSize: 20})
	if err != nil {
		return nil, err
	}
	fieldChangeLogs, _, err := s.audit.ListFieldChangeLogs(ctx, auditpkg.FieldChangeLogListInput{OrgID: orgID, ArticleID: result.ArticleID, Page: 1, PageSize: 20})
	if err != nil {
		return nil, err
	}
	return &ResultDetail{Result: *result, Hits: hits, OperationLogs: operationLogs, FieldChangeLogs: fieldChangeLogs}, nil
}
