package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

var ErrInvalidTaskInput = errors.New("invalid task input")
var ErrTaskNotFound = errors.New("task not found")
var ErrTaskDeleteNotAllowed = errors.New("task delete not allowed")

type KeywordRepository interface {
	ListByIDs(ctx context.Context, orgID uint64, ids []uint64) ([]rulespkg.KeywordRecord, map[uint64][]domainpkg.InspectionKeywordScope, error)
}

type ArticleRepository interface {
	ListCandidateArticles(ctx context.Context, filter CandidateArticleFilter) ([]scanpkg.CandidateArticle, uint64, error)
}

type ImmediateRelay interface {
	TryDispatchMessage(ctx context.Context, outboxID uint64) bool
}

type TaskService struct {
	db       *gorm.DB
	keywords KeywordRepository
	articles ArticleRepository
}

func NewTaskService(db *gorm.DB, keywords KeywordRepository, articles ArticleRepository) *TaskService {
	return &TaskService{db: db, keywords: keywords, articles: articles}
}

func (s *TaskService) List(ctx context.Context, input TaskListInput) (*TaskListResult, error) {
	if s == nil || s.db == nil || input.OrgID == 0 {
		return nil, ErrInvalidTaskInput
	}

	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	query := s.db.WithContext(ctx).Model(&domainpkg.InspectionTask{}).Where("orgid = ?", input.OrgID)

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

	items := make([]domainpkg.InspectionTask, 0, pageSize)
	if err := query.
		Order("create_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &TaskListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *TaskService) Get(ctx context.Context, orgID, taskID uint64) (*domainpkg.InspectionTask, error) {
	if s == nil || s.db == nil || orgID == 0 || taskID == 0 {
		return nil, ErrInvalidTaskInput
	}

	var task domainpkg.InspectionTask
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// Create 只负责保存任务快照与置为 pending，真正扫描由 worker 异步接手。
func (s *TaskService) Create(ctx context.Context, input CreateInspectionTaskInput) (*domainpkg.InspectionTask, error) {
	prepared, err := s.prepareTaskCreate(ctx, input)
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return persistPreparedTaskGraph(tx, prepared)
	}); err != nil {
		return nil, err
	}
	return prepared.Task, nil
}

func (s *TaskService) CreateWithOutbox(ctx context.Context, input CreateInspectionTaskInput) (*domainpkg.InspectionTask, *domainpkg.InspectionTaskOutboxMessage, error) {
	prepared, err := s.prepareTaskCreate(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	outbox := &domainpkg.InspectionTaskOutboxMessage{}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := persistPreparedTaskGraph(tx, prepared); err != nil {
			return err
		}

		prepared.Payload.TaskID = prepared.Task.ID
		payload, err := json.Marshal(prepared.Payload)
		if err != nil {
			return err
		}
		*outbox = domainpkg.InspectionTaskOutboxMessage{
			OrgID:       prepared.Task.OrgID,
			TaskID:      prepared.Task.ID,
			MessageType: domainpkg.TaskOutboxMessageTypeRunTask,
			Status:      domainpkg.TaskOutboxStatusPending,
			Payload:     string(payload),
		}
		return tx.Create(outbox).Error
	}); err != nil {
		return nil, nil, err
	}
	return prepared.Task, outbox, nil
}

func (s *TaskService) CreateAndEnqueue(ctx context.Context, input CreateInspectionTaskInput, relay ImmediateRelay) (*domainpkg.InspectionTask, error) {
	task, outbox, err := s.CreateWithOutbox(ctx, input)
	if err != nil {
		return nil, err
	}
	if relay != nil {
		relay.TryDispatchMessage(ctx, outbox.ID)
	}
	return task, nil
}

type preparedTaskCreate struct {
	Task         *domainpkg.InspectionTask
	TaskKeywords []domainpkg.InspectionTaskKeyword
	Payload      queuetasks.ArticleInspectTaskPayload
}

func (s *TaskService) prepareTaskCreate(ctx context.Context, input CreateInspectionTaskInput) (*preparedTaskCreate, error) {
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
		articleState = domainpkg.ArticleStateOnline
	}

	requestSnapshot, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	ruleSnapshot, err := json.Marshal(rulespkg.BuildKeywordRuleSnapshots(keywords, scopesByKeyword))
	if err != nil {
		return nil, err
	}

	operator := sharedpkg.ResolveOperator(ctx)
	now := time.Now().UTC()
	task := &domainpkg.InspectionTask{
		OrgID:              input.OrgID,
		TaskNo:             buildTaskNumber(now),
		Status:             domainpkg.TaskStatusPending,
		ArticleStateFilter: strconv.Itoa(int(articleState)),
		PublishTimeStart:   input.PublishTimeStart,
		PublishTimeEnd:     input.PublishTimeEnd,
		ArticleID:          input.ArticleID,
		TitleLike:          strings.TrimSpace(input.TitleLike),
		IncludeBody:        input.IncludeBody,
		RequestSnapshot:    string(requestSnapshot),
		RuleSnapshot:       string(ruleSnapshot),
		CreatorID:          operator.ID,
		CreatorName:        operator.Name,
	}

	taskKeywords := make([]domainpkg.InspectionTaskKeyword, 0, len(keywordIDs))
	for _, keywordID := range keywordIDs {
		taskKeywords = append(taskKeywords, domainpkg.InspectionTaskKeyword{OrgID: input.OrgID, KeywordID: keywordID})
	}
	return &preparedTaskCreate{
		Task:         task,
		TaskKeywords: taskKeywords,
		Payload: queuetasks.ArticleInspectTaskPayload{
			OrgID:         input.OrgID,
			TriggerSource: "api",
			OperatorID:    operator.ID,
			OperatorName:  operator.Name,
		},
	}, nil
}

func persistPreparedTaskGraph(tx *gorm.DB, prepared *preparedTaskCreate) error {
	if tx == nil || prepared == nil || prepared.Task == nil {
		return ErrInvalidTaskInput
	}
	if err := tx.Create(prepared.Task).Error; err != nil {
		return err
	}
	for index := range prepared.TaskKeywords {
		prepared.TaskKeywords[index].TaskID = prepared.Task.ID
	}
	if len(prepared.TaskKeywords) == 0 {
		return nil
	}
	return tx.Create(&prepared.TaskKeywords).Error
}

// Delete 当前只允许删除 pending / failed 任务，并会级联清理该任务写出的所有巡检数据。
func (s *TaskService) Delete(ctx context.Context, orgID, taskID uint64) error {
	if s == nil || s.db == nil || orgID == 0 || taskID == 0 {
		return ErrInvalidTaskInput
	}

	var task domainpkg.InspectionTask
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	if !taskCanBeDeleted(task.Status) {
		return ErrTaskDeleteNotAllowed
	}

	return s.deleteTaskGraph(ctx, orgID, taskID)
}

func taskCanBeDeleted(status string) bool {
	switch strings.TrimSpace(status) {
	case domainpkg.TaskStatusPending, domainpkg.TaskStatusFailed:
		return true
	default:
		return false
	}
}

func (s *TaskService) deleteTaskGraph(ctx context.Context, orgID, taskID uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&domainpkg.InspectionTaskKeyword{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&domainpkg.InspectionTaskOutboxMessage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&domainpkg.InspectionResultHit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&domainpkg.InspectionResult{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&domainpkg.InspectionOperationLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&domainpkg.InspectionFieldChangeLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("orgid = ? AND task_id = ?", orgID, taskID).Delete(&domainpkg.InspectionAction{}).Error; err != nil {
			return err
		}
		result := tx.Where("orgid = ? AND id = ?", orgID, taskID).Delete(&domainpkg.InspectionTask{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrTaskNotFound
		}
		return nil
	})
}

func buildTaskNumber(now time.Time) string {
	return fmt.Sprintf("inspect-%s", now.Format("20060102150405.000000000"))
}

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
