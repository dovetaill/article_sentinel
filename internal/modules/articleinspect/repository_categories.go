package articleinspect

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) ListOrgs(ctx context.Context) ([]ChuangqiOrg, error) {
	if r == nil || r.db == nil {
		return nil, ErrInvalidCategoryInput
	}

	items := make([]ChuangqiOrg, 0)
	if err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("sort ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CategoryRepository) List(ctx context.Context, input CategoryListInput) ([]InspectionCategory, int64, error) {
	if r == nil || r.db == nil || input.OrgID == 0 {
		return nil, 0, ErrInvalidCategoryInput
	}

	page, pageSize := normalizePage(input.Page, input.PageSize)
	query := r.db.WithContext(ctx).Model(&InspectionCategory{}).Where("orgid = ?", input.OrgID)
	if input.Enabled != nil {
		query = query.Where("enabled = ?", *input.Enabled)
	}
	if text := strings.TrimSpace(input.Query); text != "" {
		like := "%" + text + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]InspectionCategory, 0, pageSize)
	if err := query.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *CategoryRepository) Get(ctx context.Context, orgID, id uint64) (*InspectionCategory, error) {
	if r == nil || r.db == nil || orgID == 0 || id == 0 {
		return nil, ErrInvalidCategoryInput
	}

	var item InspectionCategory
	if err := r.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *CategoryRepository) Create(ctx context.Context, item *InspectionCategory) error {
	if r == nil || r.db == nil || item == nil || item.OrgID == 0 {
		return ErrInvalidCategoryInput
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *CategoryRepository) Update(ctx context.Context, item *InspectionCategory) error {
	if r == nil || r.db == nil || item == nil || item.OrgID == 0 || item.ID == 0 {
		return ErrInvalidCategoryInput
	}

	result := r.db.WithContext(ctx).Model(&InspectionCategory{}).
		Where("orgid = ? AND id = ?", item.OrgID, item.ID).
		Updates(map[string]any{
			"name":         item.Name,
			"code":         item.Code,
			"enabled":      item.Enabled,
			"sort":         item.Sort,
			"updater_id":   item.UpdaterID,
			"updater_name": item.UpdaterName,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, orgID, id uint64) error {
	if r == nil || r.db == nil || orgID == 0 || id == 0 {
		return ErrInvalidCategoryInput
	}

	result := r.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, id).Delete(&InspectionCategory{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) PatchEnabled(ctx context.Context, orgID, id uint64, enabled bool, updaterID uint64, updaterName string) error {
	if r == nil || r.db == nil || orgID == 0 || id == 0 {
		return ErrInvalidCategoryInput
	}

	result := r.db.WithContext(ctx).Model(&InspectionCategory{}).
		Where("orgid = ? AND id = ?", orgID, id).
		Updates(map[string]any{
			"enabled":      enabled,
			"updater_id":   updaterID,
			"updater_name": updaterName,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *CategoryRepository) CategoryExists(ctx context.Context, orgID, id uint64) (bool, error) {
	if r == nil || r.db == nil || orgID == 0 || id == 0 {
		return false, ErrInvalidCategoryInput
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&InspectionCategory{}).
		Where("orgid = ? AND id = ?", orgID, id).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
