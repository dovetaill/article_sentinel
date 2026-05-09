package articles

import (
	"context"
	"strings"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
)

func (r *ArticleRepository) ListCandidateArticles(ctx context.Context, filter taskspkg.CandidateArticleFilter) ([]scanpkg.CandidateArticle, uint64, error) {
	if r == nil || r.db == nil || filter.OrgID == 0 {
		return nil, 0, taskspkg.ErrInvalidTaskInput
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	state := filter.ArticleState
	if state == 0 {
		state = domainpkg.ArticleStateOnline
	}

	query := r.db.WithContext(ctx).
		Model(&articleModel{}).
		Where("orgid = ?", filter.OrgID).
		Where("state = ?", state)
	if filter.PublishTimeStart != nil {
		query = query.Where("publish_at_time >= ?", filter.PublishTimeStart.Unix())
	}
	if filter.PublishTimeEnd != nil {
		query = query.Where("publish_at_time <= ?", filter.PublishTimeEnd.Unix())
	}
	if filter.ArticleID != 0 {
		query = query.Where("id = ?", filter.ArticleID)
	}
	if titleLike := strings.TrimSpace(filter.TitleLike); titleLike != "" {
		query = query.Where("title LIKE ?", "%"+titleLike+"%")
	}
	if filter.AfterID != 0 {
		query = query.Where("id > ?", filter.AfterID)
	}

	articles := make([]articleModel, 0, limit)
	if err := query.Order("id ASC").Limit(limit).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	if len(articles) == 0 {
		return []scanpkg.CandidateArticle{}, 0, nil
	}

	bodyByArticleID, err := r.loadBodies(ctx, extractArticleIDsFromModels(articles))
	if err != nil {
		return nil, 0, err
	}

	items := make([]scanpkg.CandidateArticle, 0, len(articles))
	for _, article := range articles {
		items = append(items, scanpkg.CandidateArticle{
			ID:            article.ID,
			OrgID:         article.OrgID,
			Title:         article.Title,
			ShortTitle:    nullableStringToString(article.ShortTitle),
			RichTitle:     article.RichTitle,
			Keyword:       article.Keyword,
			Desc:          article.Desc,
			Body:          bodyByArticleID[article.ID],
			State:         article.State,
			PublishAtTime: unixSecondsPointer(article.PublishAtUnix),
		})
	}

	return items, articles[len(articles)-1].ID, nil
}
