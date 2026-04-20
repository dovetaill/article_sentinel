package articleinspect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

type Worker struct {
	db          *gorm.DB
	scanner     Scanner
	articleRepo *ArticleRepository
	batchSize   int
}

func NewWorker(db *gorm.DB) *Worker {
	return &Worker{
		db:          db,
		scanner:     NewKeywordScanner(),
		articleRepo: NewArticleRepository(db),
		batchSize:   100,
	}
}

func (w *Worker) ExecuteTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	if w == nil || w.db == nil || payload.TaskID == 0 || payload.OrgID == 0 {
		return ErrInvalidTaskInput
	}
	if w.scanner == nil {
		w.scanner = NewKeywordScanner()
	}
	if w.articleRepo == nil {
		w.articleRepo = NewArticleRepository(w.db)
	}
	if w.batchSize <= 0 {
		w.batchSize = 100
	}

	task, err := w.startTask(ctx, payload)
	if err != nil {
		return err
	}

	rules, err := decodeTaskRules(task.RuleSnapshot)
	if err != nil {
		_ = w.finishTask(ctx, task.ID, payload.OrgID, TaskStatusFailed, 0, 0, 0, 0, err.Error())
		return err
	}

	stateFilter := parseArticleStateFilter(task.ArticleStateFilter)
	var totalScanned, hitArticles, hitCount, failCount int64
	var afterID uint64

	for {
		items, nextCursor, listErr := w.articleRepo.ListCandidateArticles(ctx, CandidateArticleFilter{
			OrgID:            payload.OrgID,
			ArticleState:     stateFilter,
			PublishTimeStart: task.PublishTimeStart,
			PublishTimeEnd:   task.PublishTimeEnd,
			ArticleID:        task.ArticleID,
			TitleLike:        task.TitleLike,
			AfterID:          afterID,
			Limit:            w.batchSize,
		})
		if listErr != nil {
			_ = w.finishTask(ctx, task.ID, payload.OrgID, TaskStatusFailed, totalScanned, hitArticles, hitCount, failCount, listErr.Error())
			return listErr
		}
		if len(items) == 0 {
			break
		}

		for _, article := range items {
			totalScanned++
			hits, scanErr := w.scanner.ScanArticle(ctx, article, rules)
			if scanErr != nil {
				failCount++
				continue
			}
			if len(hits) == 0 {
				continue
			}
			hitArticles++
			hitCount += int64(len(hits))
			if persistErr := w.persistArticleResult(ctx, payload.OrgID, task.ID, article, hits); persistErr != nil {
				failCount++
			}
		}

		if len(items) < w.batchSize || nextCursor == 0 || nextCursor == afterID {
			break
		}
		afterID = nextCursor
	}

	status := resolveTaskStatus(totalScanned, failCount)
	return w.finishTask(ctx, task.ID, payload.OrgID, status, totalScanned, hitArticles, hitCount, failCount, "")
}

func (w *Worker) startTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) (*InspectionTask, error) {
	result := w.db.WithContext(ctx).Model(&InspectionTask{}).
		Where("id = ? AND orgid = ? AND status = ?", payload.TaskID, payload.OrgID, TaskStatusPending).
		Updates(map[string]any{
			"status":     TaskStatusRunning,
			"started_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("task %d is not pending", payload.TaskID)
	}

	var task InspectionTask
	if err := w.db.WithContext(ctx).Where("id = ? AND orgid = ?", payload.TaskID, payload.OrgID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

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

func (w *Worker) finishTask(ctx context.Context, taskID, orgID uint64, status string, totalScanned, hitArticles, hitCount, failCount int64, errorMessage string) error {
	updates := map[string]any{
		"status":        status,
		"total_scanned": totalScanned,
		"hit_articles":  hitArticles,
		"hit_count":     hitCount,
		"fail_count":    failCount,
		"finished_at":   time.Now().UTC(),
		"error_message": strings.TrimSpace(errorMessage),
	}
	return w.db.WithContext(ctx).Model(&InspectionTask{}).
		Where("id = ? AND orgid = ?", taskID, orgID).
		Updates(updates).Error
}

func decodeTaskRules(snapshot string) ([]KeywordRule, error) {
	if strings.TrimSpace(snapshot) == "" {
		return nil, errors.New("task rule snapshot is required")
	}

	var rules []KeywordRule
	if err := json.Unmarshal([]byte(snapshot), &rules); err == nil && len(rules) > 0 {
		return rules, nil
	}

	var dtos []KeywordDTO
	if err := json.Unmarshal([]byte(snapshot), &dtos); err != nil {
		return nil, err
	}
	rules = make([]KeywordRule, 0, len(dtos))
	for _, dto := range dtos {
		rules = append(rules, KeywordRule{
			ID:            dto.ID,
			Name:          dto.Name,
			Category:      dto.Category,
			MatchType:     dto.MatchType,
			RiskLevel:     dto.RiskLevel,
			SuggestAction: dto.SuggestAction,
			Scopes:        append([]string(nil), dto.Scopes...),
		})
	}
	return rules, nil
}

func parseArticleStateFilter(value string) int8 {
	value = strings.TrimSpace(value)
	if value == "" {
		return ArticleStateOnline
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return ArticleStateOnline
	}
	return int8(parsed)
}

func resolveTaskStatus(totalScanned, failCount int64) string {
	switch {
	case totalScanned == 0:
		return TaskStatusSuccess
	case failCount == 0:
		return TaskStatusSuccess
	case failCount >= totalScanned:
		return TaskStatusFailed
	default:
		return TaskStatusPartialSuccess
	}
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
