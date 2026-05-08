package articleinspect

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) loadBodies(ctx context.Context, articleIDs []uint64) (map[uint64]string, error) {
	result := make(map[uint64]string, len(articleIDs))
	if len(articleIDs) == 0 {
		return result, nil
	}

	infos := make([]ArticleInfo, 0, len(articleIDs))
	if err := r.db.WithContext(ctx).Where("id IN ?", articleIDs).Find(&infos).Error; err != nil {
		return nil, err
	}
	for _, info := range infos {
		result[info.ID] = info.Body
	}
	return result, nil
}

func extractArticleIDsFromModels(items []Article) []uint64 {
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func unixSecondsPointer(value int64) *time.Time {
	if value <= 0 {
		return nil
	}
	timestamp := time.Unix(value, 0).UTC()
	return &timestamp
}

func nullableStringToString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
