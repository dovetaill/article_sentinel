package articleinspect

import articlespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/articles"

var (
	ErrInvalidArticleQuery = articlespkg.ErrInvalidArticleQuery
	ErrArticleNotFound     = articlespkg.ErrArticleNotFound
)

type ArticleService = articlespkg.ArticleService

func NewArticleService(repo *articlespkg.ArticleRepository) *ArticleService {
	return articlespkg.NewArticleService(repo)
}
