package articles

import (
	"context"
	"errors"

	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

var (
	ErrInvalidArticleQuery = errors.New("invalid article query")
	ErrArticleNotFound     = errors.New("article not found")
)

type ArticleService struct {
	repo *ArticleRepository
}

func NewArticleService(repo *ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}

func (s *ArticleService) List(ctx context.Context, input ArticleListInput) (*ArticleListResult, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 {
		return nil, ErrInvalidArticleQuery
	}
	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	items, total, err := s.repo.ListArticles(ctx, input)
	if err != nil {
		return nil, err
	}
	return &ArticleListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *ArticleService) Get(ctx context.Context, orgID, articleID uint64) (*ArticleDetail, error) {
	if s == nil || s.repo == nil || orgID == 0 || articleID == 0 {
		return nil, ErrInvalidArticleQuery
	}
	return s.repo.GetArticleDetail(ctx, orgID, articleID)
}
