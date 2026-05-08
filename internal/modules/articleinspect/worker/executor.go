package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

type CandidateRepository interface {
	ListCandidateArticles(ctx context.Context, filter taskspkg.CandidateArticleFilter) ([]scanpkg.CandidateArticle, uint64, error)
}

type Executor struct {
	db          *gorm.DB
	scanner     scanpkg.Scanner
	articleRepo CandidateRepository
	batchSize   int
}

func NewWorker(db *gorm.DB) *Executor {
	return NewExecutorWithDeps(db, scanpkg.NewKeywordScanner(), newArticleRepository(db), 100)
}

func NewExecutorWithDeps(db *gorm.DB, scanner scanpkg.Scanner, articleRepo CandidateRepository, batchSize int) *Executor {
	return &Executor{db: db, scanner: scanner, articleRepo: articleRepo, batchSize: batchSize}
}

// ExecuteTask 是一期巡检主链路：拉起任务、分页扫描文稿、落结果、回写任务状态。
func (w *Executor) ExecuteTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	if w == nil || w.db == nil || payload.TaskID == 0 || payload.OrgID == 0 {
		return taskspkg.ErrInvalidTaskInput
	}
	if w.scanner == nil {
		w.scanner = scanpkg.NewKeywordScanner()
	}
	if w.articleRepo == nil {
		w.articleRepo = newArticleRepository(w.db)
	}
	if w.batchSize <= 0 {
		w.batchSize = 100
	}

	task, err := w.startTask(ctx, payload)
	if err != nil {
		return err
	}

	rules, err := DecodeTaskRules(task.RuleSnapshot)
	if err != nil {
		_ = w.finishTask(ctx, task.ID, payload.OrgID, domainpkg.TaskStatusFailed, 0, 0, 0, 0, err.Error())
		return err
	}

	stateFilter := ParseArticleStateFilter(task.ArticleStateFilter)
	var totalScanned, hitArticles, hitCount, failCount int64
	var afterID uint64

	for {
		// 按主键游标分页读取候选文稿，避免一次性扫全表。
		items, nextCursor, listErr := w.articleRepo.ListCandidateArticles(ctx, taskspkg.CandidateArticleFilter{
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
			_ = w.finishTask(ctx, task.ID, payload.OrgID, domainpkg.TaskStatusFailed, totalScanned, hitArticles, hitCount, failCount, listErr.Error())
			return listErr
		}
		if len(items) == 0 {
			break
		}

		for _, article := range items {
			totalScanned++
			// 单篇扫描失败只累计失败数，不中断整个批次。
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

	status := ResolveTaskStatus(totalScanned, failCount)
	return w.finishTask(ctx, task.ID, payload.OrgID, status, totalScanned, hitArticles, hitCount, failCount, "")
}

// startTask 只允许 pending -> running，避免重复消费同一任务。
func (w *Executor) startTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) (*domainpkg.InspectionTask, error) {
	result := w.db.WithContext(ctx).Model(&domainpkg.InspectionTask{}).
		Where("id = ? AND orgid = ? AND status = ?", payload.TaskID, payload.OrgID, domainpkg.TaskStatusPending).
		Updates(map[string]any{
			"status":     domainpkg.TaskStatusRunning,
			"started_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("task %d is not pending", payload.TaskID)
	}

	var task domainpkg.InspectionTask
	if err := w.db.WithContext(ctx).Where("id = ? AND orgid = ?", payload.TaskID, payload.OrgID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// finishTask 负责统一收口任务统计与最终状态。
func (w *Executor) finishTask(ctx context.Context, taskID, orgID uint64, status string, totalScanned, hitArticles, hitCount, failCount int64, errorMessage string) error {
	updates := map[string]any{
		"status":        status,
		"total_scanned": totalScanned,
		"hit_articles":  hitArticles,
		"hit_count":     hitCount,
		"fail_count":    failCount,
		"finished_at":   time.Now().UTC(),
		"error_message": strings.TrimSpace(errorMessage),
	}
	return w.db.WithContext(ctx).Model(&domainpkg.InspectionTask{}).
		Where("id = ? AND orgid = ?", taskID, orgID).
		Updates(updates).Error
}

// persistArticleResult 先清旧结果再写新结果，保证任务重跑时结果是幂等覆盖的。
func (w *Executor) persistArticleResult(ctx context.Context, orgID, taskID uint64, article scanpkg.CandidateArticle, hits []scanpkg.Hit) error {
	result := domainpkg.InspectionResult{
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
		DispositionStatus: domainpkg.ResultDispositionPending,
	}

	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("orgid = ? AND task_id = ? AND article_id = ?", orgID, taskID, article.ID).Delete(&domainpkg.InspectionResultHit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ? AND article_id = ?", orgID, taskID, article.ID).Delete(&domainpkg.InspectionResult{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		resultHits := make([]domainpkg.InspectionResultHit, 0, len(hits))
		for _, hit := range hits {
			resultHits = append(resultHits, domainpkg.InspectionResultHit{
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

func uniqueFieldCount(hits []scanpkg.Hit) int {
	seen := make(map[string]struct{}, len(hits))
	for _, hit := range hits {
		seen[hit.FieldName] = struct{}{}
	}
	return len(seen)
}

func uniqueKeywordCount(hits []scanpkg.Hit) int {
	seen := make(map[uint64]struct{}, len(hits))
	for _, hit := range hits {
		seen[hit.KeywordID] = struct{}{}
	}
	return len(seen)
}
