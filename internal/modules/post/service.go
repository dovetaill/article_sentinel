package post

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"unicode"
)

var (
	ErrPostNotFound      = errors.New("post not found")
	ErrDuplicatePostSlug = errors.New("post slug already exists")
	ErrInvalidPostInput  = errors.New("invalid post input")
)

type repository interface {
	Create(ctx context.Context, item *Post) error
	List(ctx context.Context, page, pageSize int) ([]Post, int64, error)
	FindByID(ctx context.Context, id uint) (*Post, error)
	FindBySlug(ctx context.Context, slug string) (*Post, error)
	Update(ctx context.Context, item *Post) error
	Delete(ctx context.Context, id uint) error
}

type Service struct {
	repo repository
}

type CreateInput struct {
	Title   string
	Slug    string
	Summary string
	Content string
}

type UpdateInput struct {
	ID      uint
	Title   string
	Slug    string
	Summary string
	Content string
}

type ListResult struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Total    int64  `json:"total"`
	Items    []Post `json:"items"`
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Post, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidPostInput
	}

	item := &Post{
		Title:   strings.TrimSpace(input.Title),
		Slug:    normalizeSlug(input.Slug, input.Title),
		Summary: strings.TrimSpace(input.Summary),
		Content: strings.TrimSpace(input.Content),
	}
	if item.Title == "" || item.Slug == "" || item.Content == "" {
		return nil, ErrInvalidPostInput
	}

	if existing, err := s.repo.FindBySlug(ctx, item.Slug); err == nil && existing != nil {
		return nil, ErrDuplicatePostSlug
	} else if err != nil && !errors.Is(err, ErrPostNotFound) {
		return nil, err
	}

	if err := s.repo.Create(ctx, item); err != nil {
		if errors.Is(err, ErrDuplicatePostSlug) {
			return nil, err
		}
		return nil, err
	}
	return item, nil
}

func (s *Service) List(ctx context.Context, page, pageSize int) (*ListResult, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidPostInput
	}

	page, pageSize = normalizePage(page, pageSize)
	items, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*Post, error) {
	if s == nil || s.repo == nil || id == 0 {
		return nil, ErrInvalidPostInput
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (*Post, error) {
	if s == nil || s.repo == nil || input.ID == 0 {
		return nil, ErrInvalidPostInput
	}

	item, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if title := strings.TrimSpace(input.Title); title != "" {
		item.Title = title
	}
	if summary := strings.TrimSpace(input.Summary); summary != "" {
		item.Summary = summary
	}
	if content := strings.TrimSpace(input.Content); content != "" {
		item.Content = content
	}
	if item.Title == "" || item.Content == "" {
		return nil, ErrInvalidPostInput
	}

	if slug := normalizeSlug(input.Slug, ""); slug != "" && slug != item.Slug {
		if existing, err := s.repo.FindBySlug(ctx, slug); err == nil && existing != nil && existing.ID != item.ID {
			return nil, ErrDuplicatePostSlug
		} else if err != nil && !errors.Is(err, ErrPostNotFound) {
			return nil, err
		}
		item.Slug = slug
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if s == nil || s.repo == nil || id == 0 {
		return ErrInvalidPostInput
	}
	return s.repo.Delete(ctx, id)
}

func StatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrDuplicatePostSlug):
		return http.StatusConflict, "post slug already exists"
	case errors.Is(err, ErrPostNotFound):
		return http.StatusNotFound, "post not found"
	case errors.Is(err, ErrInvalidPostInput):
		return http.StatusBadRequest, "invalid post input"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func normalizeSlug(slug, fallback string) string {
	source := strings.TrimSpace(slug)
	if source == "" {
		source = strings.TrimSpace(fallback)
	}
	source = strings.ToLower(source)

	var builder strings.Builder
	lastDash := false
	for _, r := range source {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		case lastDash:
			continue
		default:
			builder.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}
