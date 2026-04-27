package articleinspect

import (
	"context"
	"errors"

	"gorm.io/gorm"
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
	page, pageSize := normalizePage(input.Page, input.PageSize)
	items, total, err := s.repo.ListArticles(ctx, ArticleListInput{
		OrgID:     input.OrgID,
		Page:      page,
		PageSize:  pageSize,
		State:     input.State,
		ArticleID: input.ArticleID,
		TitleLike: input.TitleLike,
	})
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

func articleErrorFromDB(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrArticleNotFound
	}
	return err
}
