package post

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, item *Post) error {
	if r == nil || r.db == nil || item == nil {
		return ErrInvalidPostInput
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) List(ctx context.Context, page, pageSize int) ([]Post, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, ErrInvalidPostInput
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]Post, 0, pageSize)
	err := r.db.WithContext(ctx).
		Order("id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Post, error) {
	if r == nil || r.db == nil || id == 0 {
		return nil, ErrInvalidPostInput
	}

	var item Post
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Post, error) {
	if r == nil || r.db == nil || slug == "" {
		return nil, ErrInvalidPostInput
	}

	var item Post
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Update(ctx context.Context, item *Post) error {
	if r == nil || r.db == nil || item == nil || item.ID == 0 {
		return ErrInvalidPostInput
	}

	result := r.db.WithContext(ctx).Model(&Post{}).Where("id = ?", item.ID).Updates(map[string]any{
		"title":   item.Title,
		"slug":    item.Slug,
		"summary": item.Summary,
		"content": item.Content,
	}).Error
	if result != nil {
		return result
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	if r == nil || r.db == nil || id == 0 {
		return ErrInvalidPostInput
	}

	result := r.db.WithContext(ctx).Delete(&Post{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}
