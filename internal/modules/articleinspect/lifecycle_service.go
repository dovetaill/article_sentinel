package articleinspect

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type LifecycleService struct {
	db                   *gorm.DB
	actionRepo           *ActionRepository
	republishTargetState int8
}

type OfflineArticleInput struct {
	OrgID        uint64
	ArticleID    uint64
	TaskID       uint64
	ResultID     uint64
	ActionID     uint64
	OperatorID   uint64
	OperatorName string
	Reason       string
}

type UpdateArticleFieldsInput struct {
	OrgID        uint64
	ArticleID    uint64
	TaskID       uint64
	ResultID     uint64
	ActionID     uint64
	OperatorID   uint64
	OperatorName string
	Reason       string
	Fields       EditableArticleFields
}

type RepublishArticleInput struct {
	OrgID        uint64
	ArticleID    uint64
	TaskID       uint64
	ResultID     uint64
	ActionID     uint64
	OperatorID   uint64
	OperatorName string
	Reason       string
}

type LifecycleActionResult struct {
	Status      string `json:"status"`
	ArticleID   uint64 `json:"article_id"`
	BeforeState int8   `json:"before_state"`
	AfterState  int8   `json:"after_state"`
}

func NewLifecycleService(db *gorm.DB) *LifecycleService {
	return &LifecycleService{
		db:                   db,
		actionRepo:           NewActionRepository(db),
		republishTargetState: ArticleStateAuditPending,
	}
}

func (s *LifecycleService) OfflineArticle(ctx context.Context, input OfflineArticleInput) (*LifecycleActionResult, error) {
	if s == nil || s.db == nil || input.OrgID == 0 || input.ArticleID == 0 {
		return nil, ErrInvalidActionInput
	}

	var article Article
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).First(&article).Error; err != nil {
		return nil, err
	}

	result := &LifecycleActionResult{ArticleID: article.ID, BeforeState: article.State, AfterState: article.State, Status: ActionStatusSkipped}
	if article.State == ArticleStateOffline || article.State == ArticleStateOfflineSync {
		_ = s.writeOperationLog(ctx, input.OrgID, input.ActionID, input.TaskID, input.ResultID, article.ID, ActionTypeOffline, article.State, article.State, ActionStatusSkipped, input.Reason, input.OperatorID, input.OperatorName)
		return result, nil
	}

	if article.State != ArticleStateOnline {
		return nil, errors.New("article cannot be offlined from current state")
	}

	if err := s.db.WithContext(ctx).Model(&Article{}).
		Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).
		Update("state", ArticleStateOffline).Error; err != nil {
		return nil, err
	}
	result.Status = ActionStatusSuccess
	result.AfterState = ArticleStateOffline
	_ = s.writeOperationLog(ctx, input.OrgID, input.ActionID, input.TaskID, input.ResultID, article.ID, ActionTypeOffline, article.State, ArticleStateOffline, ActionStatusSuccess, input.Reason, input.OperatorID, input.OperatorName)
	return result, nil
}

func (s *LifecycleService) UpdateArticleFields(ctx context.Context, input UpdateArticleFieldsInput) ([]FieldChange, error) {
	if s == nil || s.db == nil || input.OrgID == 0 || input.ArticleID == 0 {
		return nil, ErrInvalidActionInput
	}

	var article Article
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).First(&article).Error; err != nil {
		return nil, err
	}

	var info ArticleInfo
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
		if err := tx.Model(&Article{}).
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
		if err := tx.Model(&ArticleInfo{}).
			Where("id = ?", input.ArticleID).
			Update("body", input.Fields.Body).Error; err != nil {
			return err
		}
		changeLogs := buildFieldChangeLogs(ctx, input, changes)
		if err := (&ActionRepository{db: tx}).CreateFieldChangeLogs(ctx, changeLogs); err != nil {
			return err
		}
		return s.writeOperationLogWithDB(ctx, tx, input.OrgID, input.ActionID, input.TaskID, input.ResultID, input.ArticleID, ActionTypeRectify, article.State, article.State, ActionStatusSuccess, input.Reason, input.OperatorID, input.OperatorName)
	})
	if err != nil {
		return nil, err
	}
	return changes, nil
}

func (s *LifecycleService) RepublishArticle(ctx context.Context, input RepublishArticleInput) (*LifecycleActionResult, error) {
	if s == nil || s.db == nil || input.OrgID == 0 || input.ArticleID == 0 {
		return nil, ErrInvalidActionInput
	}

	var article Article
	if err := s.db.WithContext(ctx).Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).First(&article).Error; err != nil {
		return nil, err
	}

	targetState := s.republishTargetState
	if targetState == 0 {
		targetState = ArticleStateAuditPending
	}
	if article.State == targetState {
		return &LifecycleActionResult{Status: ActionStatusSkipped, ArticleID: article.ID, BeforeState: article.State, AfterState: article.State}, nil
	}
	if err := s.db.WithContext(ctx).Model(&Article{}).
		Where("orgid = ? AND id = ?", input.OrgID, input.ArticleID).
		Update("state", targetState).Error; err != nil {
		return nil, err
	}
	_ = s.writeOperationLog(ctx, input.OrgID, input.ActionID, input.TaskID, input.ResultID, article.ID, ActionTypeRepublish, article.State, targetState, ActionStatusSuccess, input.Reason, input.OperatorID, input.OperatorName)
	return &LifecycleActionResult{Status: ActionStatusSuccess, ArticleID: article.ID, BeforeState: article.State, AfterState: targetState}, nil
}

func buildFieldChangeLogs(ctx context.Context, input UpdateArticleFieldsInput, changes []FieldChange) []InspectionFieldChangeLog {
	logs := make([]InspectionFieldChangeLog, 0, len(changes))
	for _, change := range changes {
		logs = append(logs, InspectionFieldChangeLog{
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
	enrichFieldChangeLogsWithOperator(ctx, logs)
	return logs
}

func (s *LifecycleService) writeOperationLog(ctx context.Context, orgID, actionID, taskID, resultID, articleID uint64, operationType string, beforeState, afterState int8, status, reason string, operatorID uint64, operatorName string) error {
	return s.writeOperationLogWithDB(ctx, s.db, orgID, actionID, taskID, resultID, articleID, operationType, beforeState, afterState, status, reason, operatorID, operatorName)
}

func (s *LifecycleService) writeOperationLogWithDB(ctx context.Context, db *gorm.DB, orgID, actionID, taskID, resultID, articleID uint64, operationType string, beforeState, afterState int8, status, reason string, operatorID uint64, operatorName string) error {
	if s == nil || s.actionRepo == nil {
		return nil
	}
	repo := s.actionRepo
	if db != s.db {
		repo = &ActionRepository{db: db}
	}
	logEntry := &InspectionOperationLog{
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
		OperatorID:    operatorID,
		OperatorName:  strings.TrimSpace(operatorName),
	}
	enrichOperationLogWithOperator(ctx, logEntry)
	return repo.CreateOperationLog(ctx, logEntry)
}
