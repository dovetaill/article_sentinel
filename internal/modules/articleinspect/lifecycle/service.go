package lifecycle

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	"gorm.io/gorm"
)

type auditWriter interface {
	CreateOperationLog(ctx context.Context, log *domainpkg.InspectionOperationLog) error
	CreateFieldChangeLogs(ctx context.Context, logs []domainpkg.InspectionFieldChangeLog) error
}

type LifecycleService struct {
	db                   *gorm.DB
	newAuditWriter       func(db *gorm.DB) auditWriter
	republishTargetState int8
}

func NewLifecycleService(db *gorm.DB) *LifecycleService {
	return &LifecycleService{
		db:             db,
		newAuditWriter: func(db *gorm.DB) auditWriter { return auditpkg.NewAuditRepository(db) },
		// 一期整改后的默认目标状态不是直接回 online，而是回待审。
		republishTargetState: domainpkg.ArticleStateAuditPending,
	}
}

func (s *LifecycleService) OfflineArticle(ctx context.Context, input OfflineArticleInput) (*LifecycleActionResult, error) {
	if s == nil || s.db == nil || input.OrgID == 0 || input.ArticleID == 0 {
		return nil, sharedpkg.ErrInvalidActionInput
	}

	var article domainpkg.Article
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).First(&article).Error; err != nil {
		return nil, err
	}

	result := &LifecycleActionResult{ArticleID: article.ID, BeforeState: article.State, AfterState: article.State, Status: domainpkg.ActionStatusSkipped}
	if article.State == domainpkg.ArticleStateOffline || article.State == domainpkg.ArticleStateOfflineSync {
		_ = s.writeOperationLog(
			ctx,
			input.OrgID,
			input.ActionID,
			input.TaskID,
			input.ResultID,
			article.ID,
			domainpkg.ActionTypeOffline,
			article.State,
			article.State,
			domainpkg.ActionStatusSkipped,
			input.Reason,
			input.OperatorID,
			input.OperatorName,
			auditpkg.BuildSnapshot(struct {
				OrgID         uint64 `json:"orgid"`
				TaskID        uint64 `json:"task_id,omitempty"`
				ResultID      uint64 `json:"result_id,omitempty"`
				ActionID      uint64 `json:"action_id,omitempty"`
				ArticleID     uint64 `json:"article_id"`
				OperationType string `json:"operation_type"`
				Reason        string `json:"reason,omitempty"`
			}{
				OrgID:         input.OrgID,
				TaskID:        input.TaskID,
				ResultID:      input.ResultID,
				ActionID:      input.ActionID,
				ArticleID:     article.ID,
				OperationType: domainpkg.ActionTypeOffline,
				Reason:        strings.TrimSpace(input.Reason),
			}),
		)
		return result, nil
	}

	if article.State != domainpkg.ArticleStateOnline {
		return nil, errors.New("article cannot be offlined from current state")
	}

	if err := s.db.WithContext(ctx).Model(&domainpkg.Article{}).
		Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).
		Update("state", domainpkg.ArticleStateOffline).Error; err != nil {
		return nil, err
	}
	result.Status = domainpkg.ActionStatusSuccess
	result.AfterState = domainpkg.ArticleStateOffline
	_ = s.writeOperationLog(
		ctx,
		input.OrgID,
		input.ActionID,
		input.TaskID,
		input.ResultID,
		article.ID,
		domainpkg.ActionTypeOffline,
		article.State,
		domainpkg.ArticleStateOffline,
		domainpkg.ActionStatusSuccess,
		input.Reason,
		input.OperatorID,
		input.OperatorName,
		auditpkg.BuildSnapshot(struct {
			OrgID         uint64 `json:"orgid"`
			TaskID        uint64 `json:"task_id,omitempty"`
			ResultID      uint64 `json:"result_id,omitempty"`
			ActionID      uint64 `json:"action_id,omitempty"`
			ArticleID     uint64 `json:"article_id"`
			OperationType string `json:"operation_type"`
			Reason        string `json:"reason,omitempty"`
		}{
			OrgID:         input.OrgID,
			TaskID:        input.TaskID,
			ResultID:      input.ResultID,
			ActionID:      input.ActionID,
			ArticleID:     article.ID,
			OperationType: domainpkg.ActionTypeOffline,
			Reason:        strings.TrimSpace(input.Reason),
		}),
	)
	return result, nil
}

// UpdateArticleFields 只更新允许整改的文稿字段，并同步留下字段变更日志。
func (s *LifecycleService) UpdateArticleFields(ctx context.Context, input UpdateArticleFieldsInput) ([]FieldChange, error) {
	if s == nil || s.db == nil || input.OrgID == 0 || input.ArticleID == 0 {
		return nil, sharedpkg.ErrInvalidActionInput
	}

	var article domainpkg.Article
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).First(&article).Error; err != nil {
		return nil, err
	}

	var info domainpkg.ArticleInfo
	if err := s.db.WithContext(ctx).Where("id = ?", input.ArticleID).First(&info).Error; err != nil {
		return nil, err
	}

	before := EditableArticleFields{
		Title:      article.Title,
		ShortTitle: nullableStringToString(article.ShortTitle),
		RichTitle:  article.RichTitle,
		Keyword:    article.Keyword,
		Desc:       article.Desc,
		Body:       info.Body,
	}
	changes := DiffEditableFields(before, input.Fields)
	if len(changes) == 0 {
		return nil, nil
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domainpkg.Article{}).
			Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).
			Updates(map[string]any{
				"title":       input.Fields.Title,
				"short_title": input.Fields.ShortTitle,
				"rich_title":  input.Fields.RichTitle,
				"keyword":     input.Fields.Keyword,
				"desc":        input.Fields.Desc,
			}).Error; err != nil {
			return err
		}
		if err := tx.Model(&domainpkg.ArticleInfo{}).
			Where("id = ?", input.ArticleID).
			Update("body", input.Fields.Body).Error; err != nil {
			return err
		}
		changeLogs := buildFieldChangeLogs(ctx, input, changes)
		if err := s.newAuditWriter(tx).CreateFieldChangeLogs(ctx, changeLogs); err != nil {
			return err
		}
		return s.writeOperationLogWithDB(
			ctx,
			tx,
			input.OrgID,
			input.ActionID,
			input.TaskID,
			input.ResultID,
			input.ArticleID,
			domainpkg.ActionTypeRectify,
			article.State,
			article.State,
			domainpkg.ActionStatusSuccess,
			input.Reason,
			input.OperatorID,
			input.OperatorName,
			auditpkg.BuildSnapshot(struct {
				OrgID         uint64                `json:"orgid"`
				TaskID        uint64                `json:"task_id,omitempty"`
				ResultID      uint64                `json:"result_id,omitempty"`
				ActionID      uint64                `json:"action_id,omitempty"`
				ArticleID     uint64                `json:"article_id"`
				OperationType string                `json:"operation_type"`
				Reason        string                `json:"reason,omitempty"`
				Fields        EditableArticleFields `json:"fields"`
			}{
				OrgID:         input.OrgID,
				TaskID:        input.TaskID,
				ResultID:      input.ResultID,
				ActionID:      input.ActionID,
				ArticleID:     input.ArticleID,
				OperationType: domainpkg.ActionTypeRectify,
				Reason:        strings.TrimSpace(input.Reason),
				Fields:        input.Fields,
			}),
		)
	})
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// RepublishArticle 一期默认回 audit pending，而不是直接重新上线。
func (s *LifecycleService) RepublishArticle(ctx context.Context, input RepublishArticleInput) (*LifecycleActionResult, error) {
	if s == nil || s.db == nil || input.OrgID == 0 || input.ArticleID == 0 {
		return nil, sharedpkg.ErrInvalidActionInput
	}

	var article domainpkg.Article
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).First(&article).Error; err != nil {
		return nil, err
	}

	targetState := s.republishTargetState
	if targetState == 0 {
		targetState = domainpkg.ArticleStateAuditPending
	}
	if article.State == targetState {
		return &LifecycleActionResult{Status: domainpkg.ActionStatusSkipped, ArticleID: article.ID, BeforeState: article.State, AfterState: article.State}, nil
	}
	if err := s.db.WithContext(ctx).Model(&domainpkg.Article{}).
		Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).
		Update("state", targetState).Error; err != nil {
		return nil, err
	}
	_ = s.writeOperationLog(
		ctx,
		input.OrgID,
		input.ActionID,
		input.TaskID,
		input.ResultID,
		article.ID,
		domainpkg.ActionTypeRepublish,
		article.State,
		targetState,
		domainpkg.ActionStatusSuccess,
		input.Reason,
		input.OperatorID,
		input.OperatorName,
		auditpkg.BuildSnapshot(struct {
			OrgID         uint64 `json:"orgid"`
			TaskID        uint64 `json:"task_id,omitempty"`
			ResultID      uint64 `json:"result_id,omitempty"`
			ActionID      uint64 `json:"action_id,omitempty"`
			ArticleID     uint64 `json:"article_id"`
			OperationType string `json:"operation_type"`
			Reason        string `json:"reason,omitempty"`
		}{
			OrgID:         input.OrgID,
			TaskID:        input.TaskID,
			ResultID:      input.ResultID,
			ActionID:      input.ActionID,
			ArticleID:     article.ID,
			OperationType: domainpkg.ActionTypeRepublish,
			Reason:        strings.TrimSpace(input.Reason),
		}),
	)
	return &LifecycleActionResult{Status: domainpkg.ActionStatusSuccess, ArticleID: article.ID, BeforeState: article.State, AfterState: targetState}, nil
}

func buildFieldChangeLogs(ctx context.Context, input UpdateArticleFieldsInput, changes []FieldChange) []domainpkg.InspectionFieldChangeLog {
	logs := make([]domainpkg.InspectionFieldChangeLog, 0, len(changes))
	for _, change := range changes {
		logs = append(logs, domainpkg.InspectionFieldChangeLog{
			OrgID:        input.OrgID,
			ActionID:     input.ActionID,
			TaskID:       input.TaskID,
			ResultID:     input.ResultID,
			ArticleID:    input.ArticleID,
			FieldName:    change.FieldName,
			BeforeValue:  change.BeforeValue,
			AfterValue:   change.AfterValue,
			DiffSummary:  change.DiffSummary,
			OperatorID:   input.OperatorID,
			OperatorName: strings.TrimSpace(input.OperatorName),
		})
	}
	sharedpkg.EnrichFieldChangeLogsWithOperator(ctx, logs)
	return logs
}

func (s *LifecycleService) writeOperationLog(ctx context.Context, orgID, actionID, taskID, resultID, articleID uint64, operationType string, beforeState, afterState int8, status, reason string, operatorID uint64, operatorName, requestSnapshot string) error {
	return s.writeOperationLogWithDB(ctx, s.db, orgID, actionID, taskID, resultID, articleID, operationType, beforeState, afterState, status, reason, operatorID, operatorName, requestSnapshot)
}

func (s *LifecycleService) writeOperationLogWithDB(ctx context.Context, db *gorm.DB, orgID, actionID, taskID, resultID, articleID uint64, operationType string, beforeState, afterState int8, status, reason string, operatorID uint64, operatorName, requestSnapshot string) error {
	if s == nil || s.newAuditWriter == nil {
		return nil
	}
	logEntry := &domainpkg.InspectionOperationLog{
		OrgID:         orgID,
		ActionID:      actionID,
		TaskID:        taskID,
		ResultID:      resultID,
		ArticleID:     articleID,
		OperationType: operationType,
		BeforeState:   strconv.Itoa(int(beforeState)),
		AfterState:    strconv.Itoa(int(afterState)),
		Status:        status,
		Reason:        reason,
		Summary: auditpkg.BuildOperationLogSummary(
			operationType,
			status,
			strconv.Itoa(int(beforeState)),
			strconv.Itoa(int(afterState)),
			reason,
			taskID,
			articleID,
			resultID,
		),
		RequestSnapshot: requestSnapshot,
		OperatorID:      operatorID,
		OperatorName:    strings.TrimSpace(operatorName),
	}
	sharedpkg.EnrichOperationLogWithOperator(ctx, logEntry)
	return s.newAuditWriter(db).CreateOperationLog(ctx, logEntry)
}

func nullableStringToString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
