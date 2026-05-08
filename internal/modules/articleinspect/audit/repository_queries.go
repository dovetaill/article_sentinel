package audit

import (
	"context"
	"strings"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

func (r *AuditRepository) ListOperationLogs(ctx context.Context, input OperationLogListInput) ([]domainpkg.InspectionOperationLog, int64, error) {
	if r == nil || r.db == nil || input.OrgID == 0 {
		return nil, 0, ErrInvalidLogQuery
	}
	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	query := r.db.WithContext(ctx).Model(&domainpkg.InspectionOperationLog{}).Where("orgid = ?", input.OrgID)
	if input.ArticleID != 0 {
		query = query.Where("article_id = ?", input.ArticleID)
	}
	if input.TaskID != 0 {
		query = query.Where("task_id = ?", input.TaskID)
	}
	if operator := strings.TrimSpace(input.OperatorName); operator != "" {
		query = query.Where("operator_name = ?", operator)
	}
	if input.StartAt != nil {
		query = query.Where("create_at >= ?", *input.StartAt)
	}
	if input.EndAt != nil {
		query = query.Where("create_at <= ?", *input.EndAt)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]domainpkg.InspectionOperationLog, 0, pageSize)
	if err := query.Order("create_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *AuditRepository) ListFieldChangeLogs(ctx context.Context, input FieldChangeLogListInput) ([]domainpkg.InspectionFieldChangeLog, int64, error) {
	if r == nil || r.db == nil || input.OrgID == 0 {
		return nil, 0, ErrInvalidLogQuery
	}
	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	query := r.db.WithContext(ctx).Model(&domainpkg.InspectionFieldChangeLog{}).Where("orgid = ?", input.OrgID)
	if input.ArticleID != 0 {
		query = query.Where("article_id = ?", input.ArticleID)
	}
	if fieldName := strings.TrimSpace(input.FieldName); fieldName != "" {
		query = query.Where("field_name = ?", fieldName)
	}
	if input.StartAt != nil {
		query = query.Where("create_at >= ?", *input.StartAt)
	}
	if input.EndAt != nil {
		query = query.Where("create_at <= ?", *input.EndAt)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]domainpkg.InspectionFieldChangeLog, 0, pageSize)
	if err := query.Order("create_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
