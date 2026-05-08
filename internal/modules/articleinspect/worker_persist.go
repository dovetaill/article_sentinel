package articleinspect

import (
	"context"

	"gorm.io/gorm"
)

// persistArticleResult 先清旧结果再写新结果，保证任务重跑时结果是幂等覆盖的。
func (w *Worker) persistArticleResult(ctx context.Context, orgID, taskID uint64, article CandidateArticle, hits []Hit) error {
	result := InspectionResult{
		OrgID:             orgID,
		TaskID:            taskID,
		ArticleID:         article.ID,
		ArticleTitle:      article.Title,
		ArticleState:      article.State,
		PublishAtTime:     article.PublishAtTime,
		RiskLevel:         hits[0].RiskLevel,
		SuggestAction:     hits[0].SuggestAction,
		HitFieldsCount:    int64(uniqueFieldCount(hits)),
		HitKeywordsCount:  int64(uniqueKeywordCount(hits)),
		HitCount:          int64(len(hits)),
		DispositionStatus: ResultDispositionPending,
	}

	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("orgid = ? AND task_id = ? AND article_id = ?", orgID, taskID, article.ID).Delete(&InspectionResultHit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ? AND article_id = ?", orgID, taskID, article.ID).Delete(&InspectionResult{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		resultHits := make([]InspectionResultHit, 0, len(hits))
		for _, hit := range hits {
			resultHits = append(resultHits, InspectionResultHit{
				OrgID:         orgID,
				TaskID:        taskID,
				ResultID:      result.ID,
				ArticleID:     article.ID,
				KeywordID:     hit.KeywordID,
				KeywordText:   hit.KeywordText,
				Category:      hit.Category,
				FieldName:     hit.FieldName,
				MatchType:     hit.MatchType,
				RiskLevel:     hit.RiskLevel,
				SuggestAction: hit.SuggestAction,
				MatchedText:   hit.MatchedText,
				Snippet:       hit.Snippet,
				PositionStart: int64(hit.PositionStart),
				PositionEnd:   int64(hit.PositionEnd),
			})
		}
		return tx.Create(&resultHits).Error
	})
}

func uniqueFieldCount(hits []Hit) int {
	seen := make(map[string]struct{}, len(hits))
	for _, hit := range hits {
		seen[hit.FieldName] = struct{}{}
	}
	return len(seen)
}

func uniqueKeywordCount(hits []Hit) int {
	seen := make(map[uint64]struct{}, len(hits))
	for _, hit := range hits {
		seen[hit.KeywordID] = struct{}{}
	}
	return len(seen)
}
