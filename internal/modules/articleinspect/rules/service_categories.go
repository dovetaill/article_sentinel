package rules

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

var (
	ErrInvalidCategoryInput = errors.New("invalid category input")
	ErrCategoryNotFound     = errors.New("category not found")
)

type CategoryService struct {
	repo *CategoryRepository
}

func NewCategoryService(repo *CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) ListOrgs(ctx context.Context) (*OrgListResult, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidCategoryInput
	}

	items, err := s.repo.ListOrgs(ctx)
	if err != nil {
		return nil, err
	}

	result := &OrgListResult{Items: make([]OrgDTO, 0, len(items))}
	for _, item := range items {
		result.Items = append(result.Items, OrgDTO{
			ID:      item.ID,
			Name:    item.Name,
			CateID:  item.CateID,
			Enabled: item.Enabled,
			Sort:    item.Sort,
		})
	}
	return result, nil
}

func (s *CategoryService) List(ctx context.Context, input CategoryListInput) (*CategoryListResult, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 {
		return nil, ErrInvalidCategoryInput
	}

	page, pageSize := sharedpkg.NormalizePage(input.Page, input.PageSize)
	items, total, err := s.repo.List(ctx, CategoryListInput{
		OrgID:    input.OrgID,
		Page:     page,
		PageSize: pageSize,
		Enabled:  input.Enabled,
		Query:    strings.TrimSpace(input.Query),
	})
	if err != nil {
		return nil, err
	}

	result := &CategoryListResult{Page: page, PageSize: pageSize, Total: total, Items: make([]CategoryDTO, 0, len(items))}
	for _, item := range items {
		result.Items = append(result.Items, buildCategoryDTO(item))
	}
	return result, nil
}

func (s *CategoryService) Get(ctx context.Context, orgID, id uint64) (*CategoryDTO, error) {
	if s == nil || s.repo == nil || orgID == 0 || id == 0 {
		return nil, ErrInvalidCategoryInput
	}
	item, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	dto := buildCategoryDTO(*item)
	return &dto, nil
}

func (s *CategoryService) Create(ctx context.Context, input CreateCategoryInput) (*CategoryDTO, error) {
	normalized, err := normalizeCategoryInput(input)
	if err != nil {
		return nil, err
	}
	operator := sharedpkg.ResolveOperator(ctx)
	item := &domainpkg.InspectionCategory{
		OrgID:       normalized.OrgID,
		Name:        normalized.Name,
		Enabled:     normalized.Enabled,
		Sort:        normalized.Sort,
		CreatorID:   operator.ID,
		CreatorName: operator.Name,
		UpdaterID:   operator.ID,
		UpdaterName: operator.Name,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return s.Get(ctx, item.OrgID, item.ID)
}

func (s *CategoryService) Update(ctx context.Context, input UpdateCategoryInput) (*CategoryDTO, error) {
	if input.ID == 0 {
		return nil, ErrInvalidCategoryInput
	}
	normalized, err := normalizeCategoryInput(input.CreateCategoryInput)
	if err != nil {
		return nil, err
	}
	operator := sharedpkg.ResolveOperator(ctx)
	item := &domainpkg.InspectionCategory{
		ID:          input.ID,
		OrgID:       normalized.OrgID,
		Name:        normalized.Name,
		Enabled:     normalized.Enabled,
		Sort:        normalized.Sort,
		UpdaterID:   operator.ID,
		UpdaterName: operator.Name,
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return s.Get(ctx, item.OrgID, item.ID)
}

func (s *CategoryService) Delete(ctx context.Context, orgID, id uint64) error {
	if s == nil || s.repo == nil || orgID == 0 || id == 0 {
		return ErrInvalidCategoryInput
	}
	return s.repo.Delete(ctx, orgID, id)
}

func (s *CategoryService) PatchEnabled(ctx context.Context, input PatchCategoryStatusInput) (*CategoryDTO, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 || input.CategoryID == 0 {
		return nil, ErrInvalidCategoryInput
	}
	operator := sharedpkg.ResolveOperator(ctx)
	if err := s.repo.PatchEnabled(ctx, input.OrgID, input.CategoryID, input.Enabled, operator.ID, operator.Name); err != nil {
		return nil, err
	}
	return s.Get(ctx, input.OrgID, input.CategoryID)
}

func buildCategoryDTO(item domainpkg.InspectionCategory) CategoryDTO {
	return CategoryDTO{
		ID:          item.ID,
		OrgID:       item.OrgID,
		Name:        item.Name,
		Enabled:     item.Enabled,
		Sort:        item.Sort,
		CreatorID:   item.CreatorID,
		CreatorName: item.CreatorName,
		UpdaterID:   item.UpdaterID,
		UpdaterName: item.UpdaterName,
		CreateAt:    item.CreateAt,
		UpdateAt:    item.UpdateAt,
	}
}

func normalizeCategoryInput(input CreateCategoryInput) (*CreateCategoryInput, error) {
	if input.OrgID == 0 {
		return nil, fmt.Errorf("%w: orgid is required", ErrInvalidCategoryInput)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidCategoryInput)
	}
	return &CreateCategoryInput{
		OrgID:   input.OrgID,
		Name:    name,
		Enabled: input.Enabled,
		Sort:    input.Sort,
	}, nil
}
