package results

import (
	"context"
	"strings"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

func (r *ResultRepository) ListResults(ctx context.Context, input ResultListInput) ([]ResultListItem, int64, error) {
	if r == nil || r.db == nil || input.OrgID == 0 {
		return nil, 0, ErrInvalidResultQuery
	}
	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	query := r.db.WithContext(ctx).Model(&domainpkg.InspectionResult{}).Where("orgid = ?", input.OrgID)
	if input.TaskID != 0 {
		query = query.Where("task_id = ?", input.TaskID)
	}
	if input.ArticleID != 0 {
		query = query.Where("article_id = ?", input.ArticleID)
	}
	if risk := strings.TrimSpace(input.RiskLevel); risk != "" {
		query = query.Where("risk_level = ?", risk)
	}
	if disposition := strings.TrimSpace(input.DispositionStatus); disposition != "" {
		query = query.Where("disposition_status = ?", disposition)
	}
	if title := strings.TrimSpace(input.TitleLike); title != "" {
		query = query.Where("article_title LIKE ?", "%"+title+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]domainpkg.InspectionResult, 0, pageSize)
	if err := query.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]ResultListItem, 0, len(rows))
	if len(rows) == 0 {
		return items, total, nil
	}

	resultIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		resultIDs = append(resultIDs, row.ID)
	}
	hits, err := r.listHitsByResultIDs(ctx, input.OrgID, resultIDs)
	if err != nil {
		return nil, 0, err
	}

	hitPreviewByResultID := make(map[uint64]domainpkg.InspectionResultHit, len(resultIDs))
	hitCountByResultID := make(map[uint64]int64, len(resultIDs))
	for _, hit := range hits {
		if _, ok := hitPreviewByResultID[hit.ResultID]; !ok {
			hitPreviewByResultID[hit.ResultID] = hit
		}
		hitCountByResultID[hit.ResultID]++
	}

	for _, row := range rows {
		item := ResultListItem{InspectionResult: row}
		if preview, ok := hitPreviewByResultID[row.ID]; ok {
			item.PreviewFieldName = preview.FieldName
			item.PreviewKeywordText = preview.KeywordText
			item.PreviewMatchedText = preview.MatchedText
			item.PreviewSnippet = preview.Snippet
			if count := hitCountByResultID[row.ID]; count > 1 {
				item.ExtraHitCount = count - 1
			}
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (r *ResultRepository) listHitsByResultIDs(ctx context.Context, orgID uint64, resultIDs []uint64) ([]domainpkg.InspectionResultHit, error) {
	if len(resultIDs) == 0 {
		return []domainpkg.InspectionResultHit{}, nil
	}
	items := make([]domainpkg.InspectionResultHit, 0, len(resultIDs))
	if err := r.db.WithContext(ctx).
		Where("orgid = ? AND result_id IN ?", orgID, resultIDs).
		Order("result_id ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ResultRepository) GetResult(ctx context.Context, orgID, resultID uint64) (*domainpkg.InspectionResult, error) {
	if r == nil || r.db == nil || orgID == 0 || resultID == 0 {
		return nil, ErrInvalidResultQuery
	}
	var item domainpkg.InspectionResult
	if err := r.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, resultID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ResultRepository) ListHits(ctx context.Context, orgID, resultID uint64) ([]domainpkg.InspectionResultHit, error) {
	if r == nil || r.db == nil || orgID == 0 || resultID == 0 {
		return nil, ErrInvalidResultQuery
	}
	items := make([]domainpkg.InspectionResultHit, 0)
	if err := r.db.WithContext(ctx).Where("orgid = ? AND result_id = ?", orgID, resultID).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
