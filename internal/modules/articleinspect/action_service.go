package articleinspect

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrInvalidActionInput = errors.New("invalid action input")

type BatchActionInput struct {
	OrgID        uint64
	TaskID       uint64
	ResultIDs    []uint64
	OperatorID   uint64
	OperatorName string
	Reason       string
}

type BatchActionSummary struct {
	ActionID     uint64 `json:"action_id"`
	TargetCount  int64  `json:"target_count"`
	SuccessCount int64  `json:"success_count"`
	FailCount    int64  `json:"fail_count"`
	SkipCount    int64  `json:"skip_count"`
	Status       string `json:"status"`
	ActionType   string `json:"action_type"`
}

type ActionService struct {
	db   *gorm.DB
	repo *ActionRepository
}

func NewActionService(db *gorm.DB, repo *ActionRepository) *ActionService {
	return &ActionService{db: db, repo: repo}
}

func (s *ActionService) BatchIgnore(ctx context.Context, input BatchActionInput) (*BatchActionSummary, error) {
	return s.applyDisposition(ctx, input, ActionTypeBatchIgnore, ResultDispositionIgnored)
}

func (s *ActionService) BatchProcess(ctx context.Context, input BatchActionInput) (*BatchActionSummary, error) {
	return s.applyDisposition(ctx, input, ActionTypeBatchProcess, ResultDispositionProcessed)
}

func (s *ActionService) BatchOffline(ctx context.Context, input BatchActionInput) (*BatchActionSummary, error) {
	if s == nil || s.db == nil || s.repo == nil || input.OrgID == 0 || len(input.ResultIDs) == 0 {
		return nil, ErrInvalidActionInput
	}

	now := time.Now().UTC()
	action := &InspectionAction{
		OrgID:        input.OrgID,
		ActionNo:     buildActionNumber(now),
		ActionType:   ActionTypeOffline,
		TaskID:       input.TaskID,
		TargetCount:  int64(len(input.ResultIDs)),
		Status:       ActionStatusRunning,
		Reason:       strings.TrimSpace(input.Reason),
		OperatorID:   input.OperatorID,
		OperatorName: strings.TrimSpace(input.OperatorName),
		StartedAt:    &now,
	}
	enrichActionWithOperator(ctx, action)
	if err := s.repo.CreateAction(ctx, action); err != nil {
		return nil, err
	}

	summary := &BatchActionSummary{
		ActionID:    action.ID,
		ActionType:  action.ActionType,
		TargetCount: int64(len(input.ResultIDs)),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lifecycle := NewLifecycleService(tx)
		for _, resultID := range uniqueUint64s(input.ResultIDs) {
			var result InspectionResult
			err := tx.Where("orgid = ? AND id = ?", input.OrgID, resultID).First(&result).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					summary.SkipCount++
					continue
				}
				summary.FailCount++
				continue
			}
			if result.DispositionStatus == ResultDispositionOfflined {
				summary.SkipCount++
				continue
			}

			lifecycleResult, err := lifecycle.OfflineArticle(ctx, OfflineArticleInput{
				OrgID:        input.OrgID,
				ArticleID:    result.ArticleID,
				TaskID:       input.TaskID,
				ResultID:     result.ID,
				ActionID:     action.ID,
				Reason:       input.Reason,
				OperatorID:   input.OperatorID,
				OperatorName: input.OperatorName,
			})
			if err != nil {
				summary.FailCount++
				continue
			}

			if err := tx.Model(&InspectionResult{}).
				Where("orgid = ? AND id = ?", input.OrgID, result.ID).
				Updates(map[string]any{
					"article_state":        lifecycleResult.AfterState,
					"disposition_status":   ResultDispositionOfflined,
					"latest_action_id":     action.ID,
					"latest_operator_id":   input.OperatorID,
					"latest_operator_name": strings.TrimSpace(input.OperatorName),
					"latest_action_at":     now,
				}).Error; err != nil {
				summary.FailCount++
				continue
			}

			summary.SuccessCount++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	summary.Status = ActionStatusSuccess
	if summary.FailCount > 0 && summary.SuccessCount > 0 {
		summary.Status = TaskStatusPartialSuccess
	}
	if summary.FailCount > 0 && summary.SuccessCount == 0 {
		summary.Status = ActionStatusFailed
	}
	if err := s.repo.UpdateActionSummary(ctx, action.ID, summary.Status, summary.SuccessCount, summary.FailCount, summary.SkipCount); err != nil {
		return nil, err
	}
	return summary, nil
}

func (s *ActionService) applyDisposition(ctx context.Context, input BatchActionInput, actionType, targetDisposition string) (*BatchActionSummary, error) {
	if s == nil || s.db == nil || s.repo == nil || input.OrgID == 0 || len(input.ResultIDs) == 0 {
		return nil, ErrInvalidActionInput
	}

	now := time.Now().UTC()
	action := &InspectionAction{
		OrgID:        input.OrgID,
		ActionNo:     buildActionNumber(now),
		ActionType:   actionType,
		TaskID:       input.TaskID,
		TargetCount:  int64(len(input.ResultIDs)),
		Status:       ActionStatusRunning,
		Reason:       strings.TrimSpace(input.Reason),
		OperatorID:   input.OperatorID,
		OperatorName: strings.TrimSpace(input.OperatorName),
		StartedAt:    &now,
	}
	enrichActionWithOperator(ctx, action)
	if err := s.repo.CreateAction(ctx, action); err != nil {
		return nil, err
	}

	summary := &BatchActionSummary{
		ActionID:    action.ID,
		ActionType:  actionType,
		TargetCount: int64(len(input.ResultIDs)),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, resultID := range uniqueUint64s(input.ResultIDs) {
			var result InspectionResult
			err := tx.Where("orgid = ? AND id = ?", input.OrgID, resultID).First(&result).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					summary.SkipCount++
					continue
				}
				summary.FailCount++
				continue
			}

			status := ActionStatusSuccess
			beforeDisposition := result.DispositionStatus
			if result.DispositionStatus == targetDisposition {
				status = ActionStatusSkipped
				summary.SkipCount++
			} else {
				if err := tx.Model(&InspectionResult{}).
					Where("orgid = ? AND id = ?", input.OrgID, resultID).
					Updates(map[string]any{
						"disposition_status":   targetDisposition,
						"latest_action_id":     action.ID,
						"latest_operator_id":   input.OperatorID,
						"latest_operator_name": strings.TrimSpace(input.OperatorName),
						"latest_action_at":     now,
					}).Error; err != nil {
					summary.FailCount++
					continue
				}
				summary.SuccessCount++
			}

			logEntry := &InspectionOperationLog{
				OrgID:         input.OrgID,
				ActionID:      action.ID,
				TaskID:        input.TaskID,
				ResultID:      result.ID,
				ArticleID:     result.ArticleID,
				OperationType: actionType,
				BeforeState:   beforeDisposition,
				AfterState:    targetDisposition,
				Status:        status,
				Reason:        input.Reason,
				OperatorID:    input.OperatorID,
				OperatorName:  strings.TrimSpace(input.OperatorName),
			}
			enrichOperationLogWithOperator(ctx, logEntry)
			if err := (&ActionRepository{db: tx}).CreateOperationLog(ctx, logEntry); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	summary.Status = ActionStatusSuccess
	if summary.FailCount > 0 && summary.SuccessCount > 0 {
		summary.Status = TaskStatusPartialSuccess
	}
	if summary.FailCount > 0 && summary.SuccessCount == 0 {
		summary.Status = ActionStatusFailed
	}
	if err := s.repo.UpdateActionSummary(ctx, action.ID, summary.Status, summary.SuccessCount, summary.FailCount, summary.SkipCount); err != nil {
		return nil, err
	}
	return summary, nil
}
