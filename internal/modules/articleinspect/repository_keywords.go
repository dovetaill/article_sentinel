package articleinspect

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type KeywordRepository struct {
	db *gorm.DB
}

type KeywordListFilter struct {
	OrgID    uint64
	Page     int
	PageSize int
	Enabled  *bool
	Category string
	Query    string
}

func NewKeywordRepository(db *gorm.DB) *KeywordRepository {
	return &KeywordRepository{db: db}
}

func (r *KeywordRepository) Create(ctx context.Context, keyword *InspectionKeyword, scopes []InspectionKeywordScope) error {
	if r == nil || r.db == nil || keyword == nil {
		return ErrInvalidKeywordInput
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(keyword).Error; err != nil {
			return err
		}
		if len(scopes) == 0 {
			return nil
		}
		for index := range scopes {
			scopes[index].KeywordID = keyword.ID
			scopes[index].OrgID = keyword.OrgID
		}
		return tx.Create(&scopes).Error
	})
}

func (r *KeywordRepository) Update(ctx context.Context, keyword *InspectionKeyword, scopes []InspectionKeywordScope) error {
	if r == nil || r.db == nil || keyword == nil || keyword.ID == 0 || keyword.OrgID == 0 {
		return ErrInvalidKeywordInput
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&InspectionKeyword{}).
			Where("orgid = ? AND id = ?", keyword.OrgID, keyword.ID).
			Updates(map[string]any{
				"name":           keyword.Name,
				"category":       keyword.Category,
				"match_type":     keyword.MatchType,
				"risk_level":     keyword.RiskLevel,
				"suggest_action": keyword.SuggestAction,
				"enabled":        keyword.Enabled,
				"remark":         keyword.Remark,
				"updater_id":     keyword.UpdaterID,
				"updater_name":   keyword.UpdaterName,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrKeywordNotFound
		}
		if err := tx.Where("orgid = ? AND keyword_id = ?", keyword.OrgID, keyword.ID).Delete(&InspectionKeywordScope{}).Error; err != nil {
			return err
		}
		if len(scopes) == 0 {
			return nil
		}
		for index := range scopes {
			scopes[index].KeywordID = keyword.ID
			scopes[index].OrgID = keyword.OrgID
		}
		return tx.Create(&scopes).Error
	})
}

func (r *KeywordRepository) Delete(ctx context.Context, orgID, id uint64) error {
	if r == nil || r.db == nil || orgID == 0 || id == 0 {
		return ErrInvalidKeywordInput
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("orgid = ? AND keyword_id = ?", orgID, id).Delete(&InspectionKeywordScope{}).Error; err != nil {
			return err
		}
		result := tx.Where("orgid = ? AND id = ?", orgID, id).Delete(&InspectionKeyword{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrKeywordNotFound
		}
		return nil
	})
}

func (r *KeywordRepository) Get(ctx context.Context, orgID, id uint64) (*InspectionKeyword, []InspectionKeywordScope, error) {
	if r == nil || r.db == nil || orgID == 0 || id == 0 {
		return nil, nil, ErrInvalidKeywordInput
	}

	var keyword InspectionKeyword
	err := r.db.WithContext(ctx).Where("orgid = ? AND id = ?", orgID, id).First(&keyword).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrKeywordNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	scopes, err := r.listScopes(ctx, orgID, []uint64{id})
	if err != nil {
		return nil, nil, err
	}
	return &keyword, scopes[id], nil
}

func (r *KeywordRepository) List(ctx context.Context, filter KeywordListFilter) ([]InspectionKeyword, map[uint64][]InspectionKeywordScope, int64, error) {
	if r == nil || r.db == nil || filter.OrgID == 0 {
		return nil, nil, 0, ErrInvalidKeywordInput
	}

	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	query := r.db.WithContext(ctx).Model(&InspectionKeyword{}).Where("orgid = ?", filter.OrgID)
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if category := strings.TrimSpace(filter.Category); category != "" {
		query = query.Where("category = ?", category)
	}
	if text := strings.TrimSpace(filter.Query); text != "" {
		like := "%" + text + "%"
		query = query.Where("name LIKE ? OR remark LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	items := make([]InspectionKeyword, 0, pageSize)
	if err := query.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, nil, 0, err
	}
	if len(items) == 0 {
		return items, map[uint64][]InspectionKeywordScope{}, total, nil
	}

	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	scopes, err := r.listScopes(ctx, filter.OrgID, ids)
	if err != nil {
		return nil, nil, 0, err
	}

	return items, scopes, total, nil
}

func (r *KeywordRepository) PatchEnabled(ctx context.Context, orgID, id uint64, enabled bool, updaterID uint64, updaterName string) error {
	if r == nil || r.db == nil || orgID == 0 || id == 0 {
		return ErrInvalidKeywordInput
	}

	result := r.db.WithContext(ctx).Model(&InspectionKeyword{}).
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
		return ErrKeywordNotFound
	}
	return nil
}

func (r *KeywordRepository) ListByIDs(ctx context.Context, orgID uint64, ids []uint64) ([]InspectionKeyword, map[uint64][]InspectionKeywordScope, error) {
	if r == nil || r.db == nil || orgID == 0 || len(ids) == 0 {
		return nil, nil, ErrInvalidKeywordInput
	}

	items := make([]InspectionKeyword, 0, len(ids))
	if err := r.db.WithContext(ctx).
		Where("orgid = ? AND id IN ?", orgID, ids).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return nil, nil, err
	}

	scopes, err := r.listScopes(ctx, orgID, ids)
	if err != nil {
		return nil, nil, err
	}
	return items, scopes, nil
}

func (r *KeywordRepository) listScopes(ctx context.Context, orgID uint64, keywordIDs []uint64) (map[uint64][]InspectionKeywordScope, error) {
	result := make(map[uint64][]InspectionKeywordScope, len(keywordIDs))
	if len(keywordIDs) == 0 {
		return result, nil
	}

	var scopes []InspectionKeywordScope
	if err := r.db.WithContext(ctx).
		Where("orgid = ? AND keyword_id IN ?", orgID, keywordIDs).
		Order("keyword_id ASC, scope ASC").
		Find(&scopes).Error; err != nil {
		return nil, err
	}

	for _, scope := range scopes {
		result[scope.KeywordID] = append(result[scope.KeywordID], scope)
	}
	return result, nil
}
