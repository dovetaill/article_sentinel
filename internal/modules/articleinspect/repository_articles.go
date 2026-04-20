package articleinspect

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) ListCandidateArticles(ctx context.Context, filter CandidateArticleFilter) ([]CandidateArticle, uint64, error) {
	if r == nil || r.db == nil || filter.OrgID == 0 {
		return nil, 0, ErrInvalidTaskInput
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
		state = ArticleStateOnline
	}

	query := r.db.WithContext(ctx).Model(&Article{}).
		Where("orgid = ?", filter.OrgID).
		Where("state = ?", state)
	if filter.PublishTimeStart != nil {
		query = query.Where("publish_at_time >= ?", *filter.PublishTimeStart)
	}
	if filter.PublishTimeEnd != nil {
		query = query.Where("publish_at_time <= ?", *filter.PublishTimeEnd)
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

	articles := make([]Article, 0, limit)
	if err := query.Order("id ASC").Limit(limit).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	if len(articles) == 0 {
		return []CandidateArticle{}, 0, nil
	}

	articleIDs := make([]uint64, 0, len(articles))
	for _, article := range articles {
		articleIDs = append(articleIDs, article.ID)
	}

	infos := make([]ArticleInfo, 0, len(articleIDs))
	if err := r.db.WithContext(ctx).
		Where("article_id IN ?", articleIDs).
		Find(&infos).Error; err != nil {
		return nil, 0, err
	}
	bodyByArticleID := make(map[uint64]string, len(infos))
	for _, info := range infos {
		bodyByArticleID[info.ArticleID] = info.Body
	}

	items := make([]CandidateArticle, 0, len(articles))
	for _, article := range articles {
		items = append(items, CandidateArticle{
			ID:            article.ID,
			OrgID:         article.OrgID,
			Title:         article.Title,
			ShortTitle:    article.ShortTitle,
			RichTitle:     article.RichTitle,
			Keyword:       article.Keyword,
			Desc:          article.Desc,
			Body:          bodyByArticleID[article.ID],
			State:         article.State,
			PublishAtTime: article.PublishAtTime,
		})
	}

	return items, articles[len(articles)-1].ID, nil
}
