package articleinspect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrInvalidTaskInput = errors.New("invalid task input")
var ErrTaskNotFound = errors.New("task not found")
var ErrTaskDeleteNotAllowed = errors.New("task delete not allowed")

type taskKeywordRepository interface {
	ListByIDs(ctx context.Context, orgID uint64, ids []uint64) ([]KeywordRecord, map[uint64][]InspectionKeywordScope, error)
}

type taskArticleRepository interface {
	ListCandidateArticles(ctx context.Context, filter CandidateArticleFilter) ([]CandidateArticle, uint64, error)
}

type TaskService struct {
	db       *gorm.DB
	keywords taskKeywordRepository
	articles taskArticleRepository
}

func NewTaskService(db *gorm.DB, keywords taskKeywordRepository, articles taskArticleRepository) *TaskService {
	return &TaskService{db: db, keywords: keywords, articles: articles}
}

func (s *TaskService) List(ctx context.Context, input TaskListInput) (*TaskListResult, error) {
	if s == nil || s.db == nil || input.OrgID == 0 {
		return nil, ErrInvalidTaskInput
	}

	page, pageSize := normalizePage(input.Page, input.PageSize)
	query := s.db.WithContext(ctx).Model(&InspectionTask{}).Where("orgid = ?", input.OrgID)

	if status := strings.TrimSpace(input.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if taskNo := strings.TrimSpace(input.TaskNo); taskNo != "" {
		query = query.Where("task_no LIKE ?", "%"+taskNo+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	items := make([]InspectionTask, 0, pageSize)
	if err := query.
		Order("create_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &TaskListResult{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    items,
	}, nil
}

func (s *TaskService) Get(ctx context.Context, orgID, taskID uint64) (*InspectionTask, error) {
	if s == nil || s.db == nil || orgID == 0 || taskID == 0 {
		return nil, ErrInvalidTaskInput
	}

	var task InspectionTask
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// Create 只负责保存任务快照与置为 pending，真正扫描由 worker 异步接手。
func (s *TaskService) Create(ctx context.Context, input CreateInspectionTaskInput) (*InspectionTask, error) {
	if s == nil || s.db == nil || s.keywords == nil {
		return nil, ErrInvalidTaskInput
	}
	if input.OrgID == 0 {
		return nil, fmt.Errorf("%w: orgid is required", ErrInvalidTaskInput)
	}

	keywordIDs := uniqueUint64s(input.KeywordIDs)
	if len(keywordIDs) == 0 {
		return nil, fmt.Errorf("%w: keyword_ids are required", ErrInvalidTaskInput)
	}

	keywords, scopesByKeyword, err := s.keywords.ListByIDs(ctx, input.OrgID, keywordIDs)
	if err != nil {
		return nil, err
	}
	if len(keywords) != len(keywordIDs) {
		return nil, fmt.Errorf("%w: keyword_ids not found", ErrInvalidTaskInput)
	}

	articleState := input.ArticleState
	if articleState == 0 {
		articleState = ArticleStateOnline
	}

	requestSnapshot, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	ruleSnapshot, err := json.Marshal(buildKeywordSnapshots(keywords, scopesByKeyword))
	if err != nil {
		return nil, err
	}

	operatorID, operatorName := auditOperatorFromContext(ctx)
	now := time.Now().UTC()
	task := &InspectionTask{
		OrgID:              input.OrgID,
		TaskNo:             buildTaskNumber(now),
		Status:             TaskStatusPending,
		ArticleStateFilter: strconv.Itoa(int(articleState)),
		PublishTimeStart:   input.PublishTimeStart,
		PublishTimeEnd:     input.PublishTimeEnd,
		ArticleID:          input.ArticleID,
		TitleLike:          strings.TrimSpace(input.TitleLike),
		IncludeBody:        input.IncludeBody,
		RequestSnapshot:    string(requestSnapshot),
		RuleSnapshot:       string(ruleSnapshot),
		CreatorID:          operatorID,
		CreatorName:        operatorName,
	}

	taskKeywords := make([]InspectionTaskKeyword, 0, len(keywordIDs))
	for _, keywordID := range keywordIDs {
		taskKeywords = append(taskKeywords, InspectionTaskKeyword{OrgID: input.OrgID, KeywordID: keywordID})
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for index := range taskKeywords {
			taskKeywords[index].TaskID = task.ID
		}
		return tx.Create(&taskKeywords).Error
	}); err != nil {
		return nil, err
	}
	return task, nil
}

// Delete 当前只允许删除 pending / failed 任务，并会级联清理该任务写出的所有巡检数据。
func (s *TaskService) Delete(ctx context.Context, orgID, taskID uint64) error {
	if s == nil || s.db == nil || orgID == 0 || taskID == 0 {
		return ErrInvalidTaskInput
	}

	var task InspectionTask
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	if !taskCanBeDeleted(task.Status) {
		return ErrTaskDeleteNotAllowed
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&InspectionTaskKeyword{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&InspectionResultHit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&InspectionResult{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&InspectionOperationLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&InspectionFieldChangeLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&InspectionAction{}).Error; err != nil {
			return err
		}
		result := tx.Where("orgid = ? AND id = ?", orgID, taskID).Delete(&InspectionTask{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrTaskNotFound
		}
		return nil
	})
}

// taskCanBeDeleted 明确限制可删状态，避免运行中或已成功任务被直接硬删。
func taskCanBeDeleted(status string) bool {
	switch strings.TrimSpace(status) {
	case TaskStatusPending, TaskStatusFailed:
		return true
	default:
		return false
	}
}

func buildKeywordSnapshots(keywords []KeywordRecord, scopesByKeyword map[uint64][]InspectionKeywordScope) []KeywordDTO {
	items := make([]KeywordDTO, 0, len(keywords))
	for _, keyword := range keywords {
		items = append(items, *buildKeywordDTO(&keyword, scopesByKeyword[keyword.ID]))
	}
	return items
}

func buildTaskNumber(now time.Time) string {
	return fmt.Sprintf("inspect-%s", now.Format("20060102150405.000000000"))
}

// uniqueUint64s 在任务创建和批量动作中复用，保证去重后顺序稳定，便于测试断言。
func uniqueUint64s(values []uint64) []uint64 {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
