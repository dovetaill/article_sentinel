package actions

import (
	"context"
	"strings"
	"time"

	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	lifecyclepkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/lifecycle"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	"gorm.io/gorm"
)

type ArticleLifecycle interface {
	OfflineArticle(ctx context.Context, input lifecyclepkg.OfflineArticleInput) (*lifecyclepkg.LifecycleActionResult, error)
}

type ActionService struct {
	db           *gorm.DB
	repo         *ActionRepository
	newLifecycle func(tx *gorm.DB) ArticleLifecycle
}

func NewActionService(db *gorm.DB, repo *ActionRepository, newLifecycle func(tx *gorm.DB) ArticleLifecycle) *ActionService {
	if newLifecycle == nil {
		newLifecycle = func(tx *gorm.DB) ArticleLifecycle {
			return lifecyclepkg.NewLifecycleService(tx)
		}
	}
	return &ActionService{db: db, repo: repo, newLifecycle: newLifecycle}
}

func (s *ActionService) BatchIgnore(ctx context.Context, input BatchActionInput) (*BatchActionSummary, error) {
	return s.applyDisposition(ctx, input, domainpkg.ActionTypeBatchIgnore, domainpkg.ResultDispositionIgnored)
}

func (s *ActionService) BatchProcess(ctx context.Context, input BatchActionInput) (*BatchActionSummary, error) {
	return s.applyDisposition(ctx, input, domainpkg.ActionTypeBatchProcess, domainpkg.ResultDispositionProcessed)
}

// BatchOffline 会调用生命周期服务真正修改文稿状态，而不是只改结果表处置状态。
func (s *ActionService) BatchOffline(ctx context.Context, input BatchActionInput) (*BatchActionSummary, error) {
	if s == nil || s.db == nil || s.repo == nil || input.OrgID == 0 || len(input.ResultIDs) == 0 {
		return nil, sharedpkg.ErrInvalidActionInput
	}

	now := time.Now().UTC()
	action := &domainpkg.InspectionAction{
		OrgID:       input.OrgID,
		ActionNo:    buildActionNumber(now),
		ActionType:  domainpkg.ActionTypeOffline,
		TaskID:      input.TaskID,
		TargetCount: int64(len(input.ResultIDs)),
		Status:      domainpkg.ActionStatusRunning,
		Reason:      strings.TrimSpace(input.Reason),
		RequestSnapshot: auditpkg.BuildSnapshot(struct {
			OrgID      uint64   `json:"orgid"`
			TaskID     uint64   `json:"task_id,omitempty"`
			ResultIDs  []uint64 `json:"result_ids"`
			Reason     string   `json:"reason,omitempty"`
			ActionType string   `json:"action_type"`
		}{
			OrgID:      input.OrgID,
			TaskID:     input.TaskID,
			ResultIDs:  append([]uint64(nil), input.ResultIDs...),
			Reason:     strings.TrimSpace(input.Reason),
			ActionType: domainpkg.ActionTypeOffline,
		}),
		OperatorID:   input.OperatorID,
		OperatorName: strings.TrimSpace(input.OperatorName),
		StartedAt:    &now,
	}
	sharedpkg.EnrichActionWithOperator(ctx, action)
	if err := s.repo.CreateAction(ctx, action); err != nil {
		return nil, err
	}

	summary := &BatchActionSummary{
		ActionID:    action.ID,
		ActionType:  action.ActionType,
		TargetCount: int64(len(input.ResultIDs)),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lifecycle := s.newLifecycle(tx)
		for _, resultID := range uniqueUint64s(input.ResultIDs) {
			var result domainpkg.InspectionResult
			err := tx.Where("orgid = ? AND id = ?", input.OrgID, resultID).First(&result).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					summary.SkipCount++
					continue
				}
				summary.FailCount++
				continue
			}
			if result.DispositionStatus == domainpkg.ResultDispositionOfflined {
				summary.SkipCount++
				continue
			}

			taskID := auditpkg.ResolveTaskID(input.TaskID, result.TaskID)
			lifecycleResult, err := lifecycle.OfflineArticle(ctx, lifecyclepkg.OfflineArticleInput{
				OrgID:        input.OrgID,
				ArticleID:    result.ArticleID,
				TaskID:       taskID,
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

			if err := tx.Model(&domainpkg.InspectionResult{}).
				Where("orgid = ? AND id = ?", input.OrgID, result.ID).
				Updates(map[string]any{
					"article_state":        lifecycleResult.AfterState,
					"disposition_status":   domainpkg.ResultDispositionOfflined,
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

	summary.Status = domainpkg.ActionStatusSuccess
	if summary.FailCount > 0 && summary.SuccessCount > 0 {
		summary.Status = domainpkg.TaskStatusPartialSuccess
	}
	if summary.FailCount > 0 && summary.SuccessCount == 0 {
		summary.Status = domainpkg.ActionStatusFailed
	}
	if err := s.repo.UpdateActionSummary(ctx, action.ID, summary.Status, summary.SuccessCount, summary.FailCount, summary.SkipCount); err != nil {
		return nil, err
	}
	return summary, nil
}

// applyDisposition 用于“忽略/已处理”这类只改结果处置状态、不改文稿生命周期的批量动作。
func (s *ActionService) applyDisposition(ctx context.Context, input BatchActionInput, actionType, targetDisposition string) (*BatchActionSummary, error) {
	if s == nil || s.db == nil || s.repo == nil || input.OrgID == 0 || len(input.ResultIDs) == 0 {
		return nil, sharedpkg.ErrInvalidActionInput
	}

	now := time.Now().UTC()
	action := &domainpkg.InspectionAction{
		OrgID:       input.OrgID,
		ActionNo:    buildActionNumber(now),
		ActionType:  actionType,
		TaskID:      input.TaskID,
		TargetCount: int64(len(input.ResultIDs)),
		Status:      domainpkg.ActionStatusRunning,
		Reason:      strings.TrimSpace(input.Reason),
		RequestSnapshot: auditpkg.BuildSnapshot(struct {
			OrgID      uint64   `json:"orgid"`
			TaskID     uint64   `json:"task_id,omitempty"`
			ResultIDs  []uint64 `json:"result_ids"`
			Reason     string   `json:"reason,omitempty"`
			ActionType string   `json:"action_type"`
		}{
			OrgID:      input.OrgID,
			TaskID:     input.TaskID,
			ResultIDs:  append([]uint64(nil), input.ResultIDs...),
			Reason:     strings.TrimSpace(input.Reason),
			ActionType: actionType,
		}),
		OperatorID:   input.OperatorID,
		OperatorName: strings.TrimSpace(input.OperatorName),
		StartedAt:    &now,
	}
	sharedpkg.EnrichActionWithOperator(ctx, action)
	if err := s.repo.CreateAction(ctx, action); err != nil {
		return nil, err
	}

	summary := &BatchActionSummary{
		ActionID:    action.ID,
		ActionType:  actionType,
		TargetCount: int64(len(input.ResultIDs)),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		auditWriter := auditpkg.NewAuditRepository(tx)
		for _, resultID := range uniqueUint64s(input.ResultIDs) {
			var result domainpkg.InspectionResult
			err := tx.Where("orgid = ? AND id = ?", input.OrgID, resultID).First(&result).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					summary.SkipCount++
					continue
				}
				summary.FailCount++
				continue
			}

			status := domainpkg.ActionStatusSuccess
			beforeDisposition := result.DispositionStatus
			taskID := auditpkg.ResolveTaskID(input.TaskID, result.TaskID)
			if result.DispositionStatus == targetDisposition {
				status = domainpkg.ActionStatusSkipped
				summary.SkipCount++
			} else {
				if err := tx.Model(&domainpkg.InspectionResult{}).
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

			logEntry := &domainpkg.InspectionOperationLog{
				OrgID:         input.OrgID,
				ActionID:      action.ID,
				TaskID:        taskID,
				ResultID:      result.ID,
				ArticleID:     result.ArticleID,
				OperationType: actionType,
				BeforeState:   beforeDisposition,
				AfterState:    targetDisposition,
				Status:        status,
				Reason:        input.Reason,
				Summary: auditpkg.BuildOperationLogSummary(
					actionType,
					status,
					beforeDisposition,
					targetDisposition,
					input.Reason,
					taskID,
					result.ArticleID,
					result.ID,
				),
				RequestSnapshot: auditpkg.BuildSnapshot(struct {
					OrgID         uint64   `json:"orgid"`
					TaskID        uint64   `json:"task_id,omitempty"`
					ActionID      uint64   `json:"action_id,omitempty"`
					ResultID      uint64   `json:"result_id,omitempty"`
					ResultIDs     []uint64 `json:"result_ids"`
					ArticleID     uint64   `json:"article_id,omitempty"`
					OperationType string   `json:"operation_type"`
					Status        string   `json:"status"`
					Reason        string   `json:"reason,omitempty"`
				}{
					OrgID:         input.OrgID,
					TaskID:        taskID,
					ActionID:      action.ID,
					ResultID:      result.ID,
					ResultIDs:     append([]uint64(nil), input.ResultIDs...),
					ArticleID:     result.ArticleID,
					OperationType: actionType,
					Status:        status,
					Reason:        strings.TrimSpace(input.Reason),
				}),
				OperatorID:   input.OperatorID,
				OperatorName: strings.TrimSpace(input.OperatorName),
			}
			sharedpkg.EnrichOperationLogWithOperator(ctx, logEntry)
			if err := auditWriter.CreateOperationLog(ctx, logEntry); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	summary.Status = domainpkg.ActionStatusSuccess
	if summary.FailCount > 0 && summary.SuccessCount > 0 {
		summary.Status = domainpkg.TaskStatusPartialSuccess
	}
	if summary.FailCount > 0 && summary.SuccessCount == 0 {
		summary.Status = domainpkg.ActionStatusFailed
	}
	if err := s.repo.UpdateActionSummary(ctx, action.ID, summary.Status, summary.SuccessCount, summary.FailCount, summary.SkipCount); err != nil {
		return nil, err
	}
	return summary, nil
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
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j] < result[i] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}
