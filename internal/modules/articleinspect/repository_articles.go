package articleinspect

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

type articleInspectSummary struct {
	ArticleID           uint64
	LatestResultID      uint64
	LatestTaskID        uint64
	LatestRiskLevel     string
	LatestSuggestAction string
	LatestDisposition   string
	LatestOperatorName  string
	LatestActionAt      *time.Time
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

	articles := make([]Article, 0, limit)
	if err := query.Order("id ASC").Limit(limit).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	if len(articles) == 0 {
		return []CandidateArticle{}, 0, nil
	}

	bodyByArticleID, err := r.loadBodies(ctx, extractArticleIDsFromModels(articles))
	if err != nil {
		return nil, 0, err
	}

	items := make([]CandidateArticle, 0, len(articles))
	for _, article := range articles {
		items = append(items, CandidateArticle{
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

func (r *ArticleRepository) ListArticles(ctx context.Context, input ArticleListInput) ([]ArticleListItem, int64, error) {
	if r == nil || r.db == nil || input.OrgID == 0 {
		return nil, 0, ErrInvalidArticleQuery
	}

	page, pageSize := normalizePage(input.Page, input.PageSize)
	query := r.db.WithContext(ctx).Model(&Article{}).Where("orgid = ?", input.OrgID)
	if input.State != nil {
		query = query.Where("state = ?", *input.State)
	}
	if text := strings.TrimSpace(input.Query); text != "" {
		like := "%" + text + "%"
		query = query.Where("title LIKE ? OR keyword LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	articles := make([]Article, 0, pageSize)
	if err := query.Order("publish_at_time DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	if len(articles) == 0 {
		return []ArticleListItem{}, total, nil
	}

	summaries, err := r.loadLatestInspectSummaries(ctx, input.OrgID, extractArticleIDsFromModels(articles))
	if err != nil {
		return nil, 0, err
	}

	items := make([]ArticleListItem, 0, len(articles))
	for _, article := range articles {
		summary := summaries[article.ID]
		items = append(items, ArticleListItem{
			ID:                  article.ID,
			OrgID:               article.OrgID,
			Title:               article.Title,
			Thumbnail:           article.Thumbnail,
			State:               article.State,
			PublishAtTime:       unixSecondsPointer(article.PublishAtUnix),
			LatestRiskLevel:     summary.LatestRiskLevel,
			LatestTaskID:        summary.LatestTaskID,
			LatestResultID:      summary.LatestResultID,
			LatestSuggestAction: summary.LatestSuggestAction,
			LatestDisposition:   summary.LatestDisposition,
			LatestOperatorName:  summary.LatestOperatorName,
			LatestActionAt:      summary.LatestActionAt,
		})
	}
	return items, total, nil
}

func (r *ArticleRepository) GetArticleDetail(ctx context.Context, orgID, articleID uint64) (*ArticleDetail, error) {
	if r == nil || r.db == nil || orgID == 0 || articleID == 0 {
		return nil, ErrInvalidArticleQuery
	}

	var article Article
	if err := r.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, articleID).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}

	bodyByArticleID, err := r.loadBodies(ctx, []uint64{articleID})
	if err != nil {
		return nil, err
	}
	summaries, err := r.loadLatestInspectSummaries(ctx, orgID, []uint64{articleID})
	if err != nil {
		return nil, err
	}
	summary := summaries[articleID]

	return &ArticleDetail{
		ID:                  article.ID,
		OrgID:               article.OrgID,
		Title:               article.Title,
		ShortTitle:          nullableStringToString(article.ShortTitle),
		RichTitle:           article.RichTitle,
		Keyword:             article.Keyword,
		Desc:                article.Desc,
		Body:                bodyByArticleID[articleID],
		Thumbnail:           article.Thumbnail,
		State:               article.State,
		PublishAtTime:       unixSecondsPointer(article.PublishAtUnix),
		LatestRiskLevel:     summary.LatestRiskLevel,
		LatestTaskID:        summary.LatestTaskID,
		LatestResultID:      summary.LatestResultID,
		LatestSuggestAction: summary.LatestSuggestAction,
		LatestDisposition:   summary.LatestDisposition,
		LatestOperatorName:  summary.LatestOperatorName,
		LatestActionAt:      summary.LatestActionAt,
	}, nil
}

func (r *ArticleRepository) loadBodies(ctx context.Context, articleIDs []uint64) (map[uint64]string, error) {
	result := make(map[uint64]string, len(articleIDs))
	if len(articleIDs) == 0 {
		return result, nil
	}

	infos := make([]ArticleInfo, 0, len(articleIDs))
	if err := r.db.WithContext(ctx).Where("id IN ?", articleIDs).Find(&infos).Error; err != nil {
		return nil, err
	}
	for _, info := range infos {
		result[info.ID] = info.Body
	}
	return result, nil
}

func (r *ArticleRepository) loadLatestInspectSummaries(ctx context.Context, orgID uint64, articleIDs []uint64) (map[uint64]articleInspectSummary, error) {
	result := make(map[uint64]articleInspectSummary, len(articleIDs))
	if len(articleIDs) == 0 {
		return result, nil
	}

	rows := make([]InspectionResult, 0)
	if err := r.db.WithContext(ctx).
		Where("orgid = ? AND article_id IN ?", orgID, articleIDs).
		Order("article_id ASC, latest_action_at DESC, task_id DESC, id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		if _, ok := result[row.ArticleID]; ok {
			continue
		}
		result[row.ArticleID] = articleInspectSummary{
			ArticleID:           row.ArticleID,
			LatestResultID:      row.ID,
			LatestTaskID:        row.TaskID,
			LatestRiskLevel:     row.RiskLevel,
			LatestSuggestAction: row.SuggestAction,
			LatestDisposition:   row.DispositionStatus,
			LatestOperatorName:  row.LatestOperatorName,
			LatestActionAt:      row.LatestActionAt,
		}
	}
	return result, nil
}

func extractArticleIDsFromModels(items []Article) []uint64 {
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func unixSecondsPointer(value int64) *time.Time {
	if value <= 0 {
		return nil
	}
	timestamp := time.Unix(value, 0).UTC()
	return &timestamp
}

func nullableStringToString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
