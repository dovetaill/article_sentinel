package post

import (
	"context"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/article-sentinel/internal/api/response"
)

type routeService interface {
	Create(ctx context.Context, input CreateInput) (*Post, error)
	List(ctx context.Context, page, pageSize int) (*ListResult, error)
	Get(ctx context.Context, id uint) (*Post, error)
	Update(ctx context.Context, input UpdateInput) (*Post, error)
	Delete(ctx context.Context, id uint) error
}

type postCreateBody struct {
	Title   string `json:"title"`
	Slug    string `json:"slug,omitempty"`
	Summary string `json:"summary"`
	Content string `json:"content"`
}

type postCreateRequest struct {
	Body postCreateBody
}

type postListRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type postIDRequest struct {
	ID string `path:"id"`
}

type postUpdateBody struct {
	Title   *string `json:"title,omitempty"`
	Slug    *string `json:"slug,omitempty"`
	Summary *string `json:"summary,omitempty"`
	Content *string `json:"content,omitempty"`
}

type postUpdateRequest struct {
	ID   string `path:"id"`
	Body postUpdateBody
}

type postEnvelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

func RegisterRoutes(api huma.API, service routeService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "post-create", Method: http.MethodPost, Path: "/api/v1/posts", Summary: "create post"}, func(ctx context.Context, input *postCreateRequest) (*postEnvelopeOutput, error) {
		item, err := service.Create(ctx, CreateInput{
			Title:   input.Body.Title,
			Slug:    input.Body.Slug,
			Summary: input.Body.Summary,
			Content: input.Body.Content,
		})
		if err != nil {
			status, message := StatusFromError(err)
			return &postEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &postEnvelopeOutput{Status: http.StatusCreated, Body: response.OK("post created", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "post-list", Method: http.MethodGet, Path: "/api/v1/posts", Summary: "list posts"}, func(ctx context.Context, input *postListRequest) (*postEnvelopeOutput, error) {
		result, err := service.List(ctx, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &postEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &postEnvelopeOutput{Status: http.StatusOK, Body: response.Paged("post list", result.Page, result.PageSize, result.Total, result.Items)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "post-get", Method: http.MethodGet, Path: "/api/v1/posts/{id}", Summary: "get post"}, func(ctx context.Context, input *postIDRequest) (*postEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &postEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid post input")}, nil
		}
		item, err := service.Get(ctx, id)
		if err != nil {
			status, message := StatusFromError(err)
			return &postEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &postEnvelopeOutput{Status: http.StatusOK, Body: response.OK("post detail", item)}, nil
	})

	postUpdateHandler := func(ctx context.Context, input *postUpdateRequest) (*postEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &postEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid post input")}, nil
		}
		item, err := service.Update(ctx, UpdateInput{
			ID:      id,
			Title:   stringValue(input.Body.Title),
			Slug:    stringValue(input.Body.Slug),
			Summary: stringValue(input.Body.Summary),
			Content: stringValue(input.Body.Content),
		})
		if err != nil {
			status, message := StatusFromError(err)
			return &postEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &postEnvelopeOutput{Status: http.StatusOK, Body: response.OK("post updated", item)}, nil
	}

	huma.Register(api, huma.Operation{OperationID: "post-update", Method: http.MethodPut, Path: "/api/v1/posts/{id}", Summary: "update post"}, postUpdateHandler)
	huma.Register(api, huma.Operation{OperationID: "post-update-patch", Method: http.MethodPatch, Path: "/api/v1/posts/{id}", Summary: "patch post"}, postUpdateHandler)

	huma.Register(api, huma.Operation{OperationID: "post-delete", Method: http.MethodDelete, Path: "/api/v1/posts/{id}", Summary: "delete post"}, func(ctx context.Context, input *postIDRequest) (*postEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &postEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid post input")}, nil
		}
		if err := service.Delete(ctx, id); err != nil {
			status, message := StatusFromError(err)
			return &postEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &postEnvelopeOutput{Status: http.StatusOK, Body: response.OK("post deleted", map[string]uint{"id": id})}, nil
	})
}

func parseID(raw string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
