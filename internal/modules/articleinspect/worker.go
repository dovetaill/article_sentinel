package articleinspect

import (
	"context"
	"fmt"
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

// NewWorker 构建巡检异步执行器。
func NewWorker(db *gorm.DB) *Worker {
	return &Worker{
		db:          db,
		scanner:     NewKeywordScanner(),
		articleRepo: NewArticleRepository(db),
		batchSize:   100,
	}
}

// ExecuteTask 是一期巡检主链路：拉起任务、分页扫描文稿、落结果、回写任务状态。
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
		// 按主键游标分页读取候选文稿，避免一次性扫全表。
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

	status := resolveTaskStatus(totalScanned, failCount)
	return w.finishTask(ctx, task.ID, payload.OrgID, status, totalScanned, hitArticles, hitCount, failCount, "")
}

// startTask 只允许 pending -> running，避免重复消费同一任务。
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

// finishTask 负责统一收口任务统计与最终状态。
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
